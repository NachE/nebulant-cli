// Nebulant
// Copyright (C) 2023 Develatio Technologies S.L.

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

import (
	"io"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// ActionContext struct
type ActionContext struct {
	Rehearsal bool
	HClient   *hcloud.Client
	Action    *blueprint.Action
	Store     base.IStore
	Logger    base.ILogger
}

func UnmarshallHCloudToSchema(response *hcloud.Response, v interface{}) error {
	var body []byte
	body, err := io.ReadAll(response.Response.Body)
	if err != nil {
		return err
	}
	err = util.UnmarshalValidJSON(body, v)
	if err != nil {
		return err
	}
	return nil
}

var NewActionContext = func(client *hcloud.Client, action *blueprint.Action, store base.IStore, logger base.ILogger) *ActionContext {
	l := logger.Duplicate()
	l.SetActionID(action.ActionID)
	return &ActionContext{
		HClient: client,
		Action:  action,
		Store:   store,
		Logger:  l,
	}
}

// ActionFunc func
type ActionFunc func(ctx *ActionContext) (*base.ActionOutput, error)

type NextType int

const (
	// NextOKKO const 0
	NextOKKO NextType = iota
	// NextOK const 1
	NextOK
	// NextKO const
	NextKO
)

type ActionLayout struct {
	F ActionFunc
	N NextType
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"create_floating_ip":   {F: CreateFloatingIP, N: NextOKKO},
	"delete_floating_ip":   {F: DeleteFloatingIP, N: NextOKKO},
	"find_floating_ips":    {F: FindFloatingIPs, N: NextOKKO},
	"findone_floating_ip":  {F: FindOneFloatingIP, N: NextOKKO},
	"assign_floating_ip":   {F: AssignFloatingIP, N: NextOKKO},
	"unassign_floating_ip": {F: UnassignFloatingIP, N: NextOKKO},

	"find_images":   {F: FindImages, N: NextOKKO},
	"findone_image": {F: FindOneImage, N: NextOKKO},
	"delete_image":  {F: FindOneImage, N: NextOKKO},

	"create_server":  {F: CreateServer, N: NextOKKO},
	"delete_server":  {F: DeleteServer, N: NextOKKO},
	"find_servers":   {F: FindServers, N: NextOKKO},
	"findone_server": {F: FindOneServer, N: NextOKKO},
	"start_server":   {F: PowerOnServer, N: NextOKKO},  // poweron server
	"stop_server":    {F: PowerOffServer, N: NextOKKO}, // poweroff server

	"create_network":  {F: CreateNetwork, N: NextOKKO},
	"delete_network":  {F: DeleteNetwork, N: NextOKKO},
	"find_networks":   {F: FindNetworks, N: NextOKKO},
	"findone_network": {F: FindOneNetwork, N: NextOKKO},

	"create_volume":  {F: CreateVolume, N: NextOKKO},
	"delete_volume":  {F: DeleteVolume, N: NextOKKO},
	"find_volumes":   {F: FindVolumes, N: NextOKKO},
	"findone_volume": {F: FindOneVolume, N: NextOKKO},
	"attach_volume":  {F: AttachVolume, N: NextOKKO},
	"detach_volume":  {F: DetachVolume, N: NextOKKO},

	"find_datacenters":   {F: FindDatacenters, N: NextOKKO},
	"findone_datacenter": {F: FindOneDatacenter, N: NextOKKO},

	"create_firewall":                {F: CreateFirewall, N: NextOKKO},
	"delete_firewall":                {F: DeleteFirewall, N: NextOKKO},
	"find_firewalls":                 {F: FindFirewalls, N: NextOKKO},
	"findone_firewall":               {F: FindOneFirewall, N: NextOKKO},
	"apply_to_resources_firewall":    {F: ApplyToResourcesFirewall, N: NextOKKO},
	"remove_from_resources_firewall": {F: RemoveFromResourcesFirewall, N: NextOKKO},
	"set_rules_firewall":             {F: SetRulesFirewall, N: NextOKKO},

	"find_isos":   {F: FindISOs, N: NextOKKO},
	"findone_iso": {F: FindOneISO, N: NextOKKO},

	"create_load_balancer":               {F: CreateLoadBalancer, N: NextOKKO},
	"delete_load_balancer":               {F: DeleteLoadBalancer, N: NextOKKO},
	"find_load_balancers":                {F: FindLoadBalancers, N: NextOKKO},
	"findone_load_balancer":              {F: FindOneLoadBalancer, N: NextOKKO},
	"attach_to_network_load_balancer":    {F: AttachToNetworkLoadBalancer, N: NextOKKO},
	"dettach_from_network_load_balancer": {F: DetachFromNetworkLoadBalancer, N: NextOKKO},

	"find_locations":   {F: FindLocations, N: NextOKKO},
	"findone_location": {F: FindOneLocation, N: NextOKKO},

	"create_primary_ip":   {F: CreatePrimaryIP, N: NextOKKO},
	"delete_primary_ip":   {F: DeletePrimaryIP, N: NextOKKO},
	"find_primary_ips":    {F: FindPrimaryIPs, N: NextOKKO},
	"findone_primary_ip":  {F: FindOnePrimaryIP, N: NextOKKO},
	"assign_primary_ip":   {F: AssignPrimaryIP, N: NextOKKO},
	"unassign_primary_ip": {F: UnassignPrimaryIP, N: NextOKKO},

	"create_ssh_key":  {F: CreateSSHKey, N: NextOKKO},
	"delete_ssh_key":  {F: DeleteSSHKey, N: NextOKKO},
	"find_ssh_keys":   {F: FindSSHKeys, N: NextOKKO},
	"findone_ssh_key": {F: FindOneSSHKey, N: NextOKKO},

	"add_target":    {F: AddTargetToLoadBalancer, N: NextOKKO},
	"remove_target": {F: RemoveTargetToLoadBalancer, N: NextOKKO},
}

// GenericHCloudOutput unmarshall response into v and return ActionContext with
// the result
func GenericHCloudOutput(ctx *ActionContext, response *hcloud.Response, v interface{}) (*base.ActionOutput, error) {
	err := UnmarshallHCloudToSchema(response, v)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, v, nil)
	return aout, nil
}
