// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package actors

// Considerations:
// - Only one instance of runActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of runActor should not be
// implemented.
//

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

type BodyType string
type PartType string

const (
	BodyTypeNone               BodyType = "none"
	BodyTypeFormData           BodyType = "form-data"
	BodyTypeXWWWFormUrlencoded BodyType = "x-www-form-urlencoded"
	BodyTypeRaw                BodyType = "raw"
	BodyTypeBinary             BodyType = "binary"
)

const (
	PartTypeText PartType = "text"
	PartTypeFile PartType = "file"
)

type httpBodyMultiPart struct {
	Name        *string  `json:"name" validate:"required"`
	Value       *string  `json:"value" validate:"required"`
	PType       PartType `json:"type" validate:"required"`
	ContentType *string  `json:"content_type"`
}

type httpBodyUrlencoded struct {
	Name  *string `json:"name" validate:"required"`
	Value *string `json:"value" validate:"required"`
}

type httpHeader struct {
	Key   *string `json:"name" validate:"required"`
	Value *string `json:"value" validate:"required"`
}

// issue #11
type httpRequestParameters struct {
	Method   *string       `json:"http_verb" validate:"required"`
	Url      *string       `json:"endpoint" validate:"required"`
	Headers  []*httpHeader `json:"headers"`
	BodyType BodyType      `json:"body_type" validate:"required"`
}

type httpRequestParametersMultiPartBody struct {
	MultipartBody []*httpBodyMultiPart `json:"body" validate:"required"`
}

type httpRequestParametersUrlEncodedBody struct {
	UrlEncodedBody []*httpBodyUrlencoded `json:"body" validate:"required"`
}

type httpRequestParametersRawBody struct {
	RawBody *string `json:"body" validate:"required"`
}

type httpRequestParametersBinaryBody struct {
	BinaryBody *string `json:"body" validate:"required"`
}

type httpRequestOutput struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Headers    string `json:"headers"`
	Body       string `json:"body"`
}

// RunRemoteScript func
func HttpRequest(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	var req *http.Request

	p := &httpRequestParameters{}
	if err = json.Unmarshal(ctx.Action.Parameters, p); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	// issue #11
	switch p.BodyType {
	case BodyTypeFormData:
		body := new(bytes.Buffer)
		w := multipart.NewWriter(body)
		defer w.Close()
		// body part definitions
		// body is:
		// {
		// 	name: "campo1",
		// 	value: "valor1",
		// 	type: "", // "text" \ "file"
		//		// "type:text" -> content_type is the value of "Content-Type" part header
		//		// "type:file" -> content_type is ignored. The path of file is in the value attr
		// content_type: "",
		// }
		param := &httpRequestParametersMultiPartBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}

		// loop for body parts
		for _, part := range param.MultipartBody {
			switch part.PType {
			case PartTypeFile: // type file: read file and write content
				file, err := os.Open(*part.Value)
				if err != nil {
					return nil, err
				}
				defer file.Close()
				finf, err := file.Stat()
				if err != nil {
					return nil, err
				}
				fcontent, err := ioutil.ReadAll(file)
				if err != nil {
					return nil, err
				}
				ff, err := w.CreateFormFile(*part.Name, finf.Name())
				if err != nil {
					return nil, err
				}
				_, err = ff.Write(fcontent)
				if err != nil {
					return nil, err
				}
				ctx.Logger.LogDebug("Append file part: " + finf.Name())
			case PartTypeText: // type text, write as part and override content_type
				h := make(textproto.MIMEHeader)
				n := strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(*part.Name)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, n))
				if part.ContentType != nil {
					h.Set("Content-Type", *part.ContentType)
				} else {
					h.Set("Content-Type", "text/plain")
				}
				ff, err := w.CreatePart(h)
				if err != nil {
					return nil, err
				}
				_, err = ff.Write([]byte(*part.Value))
				if err != nil {
					return nil, err
				}
				ctx.Logger.LogDebug("Append text content: " + *part.Value)
			}
		}
		// request with formatted body parts
		// hard debug :P
		// fmt.Println(body)
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "multipart/form-data")
	case BodyTypeNone:
		// no body with type none
		req, err = http.NewRequest(*p.Method, *p.Url, http.NoBody)
		if err != nil {
			return nil, err
		}
	case BodyTypeXWWWFormUrlencoded:
		// body part definitions
		// body is {name: "campo1", value: "valor1"}
		param := &httpRequestParametersUrlEncodedBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		// append key:value
		fdata := url.Values{}
		for _, kv := range param.UrlEncodedBody {
			fdata.Add(*kv.Name, *kv.Value)
			ctx.Logger.LogDebug("Append urlenc key: " + *kv.Name)
		}
		// url-encoded reader
		body := strings.NewReader(fdata.Encode())
		// request with url-encoded body
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case BodyTypeRaw:
		param := &httpRequestParametersRawBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		body := strings.NewReader(*param.RawBody)
		// the content type should be setted by the user
		// in httpRequestParameters.Headers
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
	case BodyTypeBinary:
		param := &httpRequestParametersBinaryBody{}
		// the body contains the path of a file
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		file, err := os.Open(*param.BinaryBody)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		fcontent, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		// detect the file mime type
		ct := http.DetectContentType(fcontent)
		body := bytes.NewBuffer(fcontent)
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", ct)
	default:
		// default is like none type
		req, err = http.NewRequest(*p.Method, *p.Url, http.NoBody)
		if err != nil {
			return nil, err
		}
	}

	// set headers
	for _, hh := range p.Headers {
		if req.Header.Get(*hh.Key) != "" {
			// already seted header
			continue
		}
		err := ctx.Store.Interpolate(hh.Value)
		if err != nil {
			return nil, err
		}
		ctx.Logger.LogDebug("Use request header, " + *hh.Key + ": " + *hh.Value)
		req.Header.Set(*hh.Key, *hh.Value)
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := httpRequestOutput{}

	// debug status
	ctx.Logger.LogDebug("Request response status: " + resp.Status)
	result.Status = resp.Status
	result.StatusCode = resp.StatusCode

	// debug headers
	sw := new(strings.Builder)
	err = resp.Header.Write(sw)
	if err != nil {
		return nil, err
	}
	ctx.Logger.LogDebug("Headers: " + sw.String())
	result.Headers = sw.String()

	// debug body
	if resp.ContentLength > 0 {
		var rawbody io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			rawbody, err = gzip.NewReader(resp.Body)
			if err != nil {
				return nil, err
			}
			defer rawbody.Close()
		default:
			rawbody = resp.Body
		}

		sw := new(strings.Builder)
		_, err = io.Copy(sw, rawbody)
		if err != nil {
			return nil, err
		}
		ctx.Logger.LogDebug("Body: " + sw.String())
		result.Body = sw.String()
	} else {
		ctx.Logger.LogDebug("Empty body")
		result.Body = ""
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}
