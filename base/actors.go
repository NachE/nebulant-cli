// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package base

import (
	"io"

	"github.com/develatio/nebulant-cli/blueprint"
)

type ActionContextRunStatus int

type ContextType int

const (
	ContextTypeRegular ContextType = iota // one parent, one child
	ContextTypeThread                     // one parent, many children
	ContextTypeJoin                       // many parents, one child
)

const (
	RunStatusReady ActionContextRunStatus = iota
	RunStatusArranging
	RunStatusRunning
	RunStatusDone
)

// ActionContext interface
type IActionContext interface {
	Type() ContextType
	Parent(IActionContext)
	Parents() []IActionContext
	Child(IActionContext)
	Children() []IActionContext
	IsThreadPoint() bool
	IsJoinPoint() bool
	GetAction() *blueprint.Action
	SetStore(IStore)
	GetStore() IStore
	// SetProvider(IProvider)
	// GetProvider() IProvider
	Done() <-chan struct{}
	WithCancelCause()
	WithEventListener(*EventListener)
	EventListener() *EventListener
	Cancel(error)
	WithRunFunc(func() (*ActionOutput, error))
	RunAction() (*ActionOutput, error)
	SetRunStatus(ActionContextRunStatus)
	GetRunStatus() ActionContextRunStatus
	//
	GetSluvaFD() io.ReadWriteCloser
	GetMustarFD() io.ReadWriteCloser
	//
	// the func that runs on DebugInit call
	WithDebugInitFunc(func())
	// returns the fd in wich the action can
	// read/write to interact with  user,
	// commonly the sluva FD of vpty
	DebugInit()
}

// IActor interface
type IActor interface {
	RunAction(action *blueprint.Action) (*ActionOutput, error)
}

// Actor struct
type Actor struct {
	Provider IProvider
}

// ActionOutput struct
type ActionOutput struct {
	Action  *blueprint.Action
	Records []*StorageRecord
}

// NewActionOutput func.
func NewActionOutput(action *blueprint.Action, storageRecordValue interface{}, storageRecordValueID *string) *ActionOutput {
	aout := &ActionOutput{
		Action: action,
	}
	record := &StorageRecord{
		Action:    action,
		RawSource: storageRecordValue,
		Value:     storageRecordValue,
		// Recursive reference. Inside aout there is also
		// reference to this StorageRecord object.
		Aout: aout,
	}
	if storageRecordValueID != nil {
		record.ValueID = *storageRecordValueID
	}
	if action.Output != nil {
		record.RefName = *action.Output
	}
	aout.Records = append(aout.Records, record)
	return aout
}
