// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package diceyml

var (
	notfoundJob                = errortype("not found job in yaml")
	notfoundService            = errortype("not found service in yaml")
	invalidService             = errortype("invalid service defined in yaml")
	emptyServiceJobList        = errortype("empty service and job list")
	notfoundVersion            = errortype("not found version in yaml")
	invalidReplicas            = errortype("invalid replicas defined in yaml")
	invalidPolicy              = errortype("invalid policy defined in yaml")
	invalidCPU                 = errortype("invalid cpu defined in yaml")
	invalidMaxCPU              = errortype("invalid max cpu defined in yaml")
	invalidMem                 = errortype("invalid memory defined in yaml")
	invalidDisk                = errortype("invalid disk defined in yaml")
	invalidNetworkMode         = errortype("invalid network mode defined in yaml, must be 'container' or 'host'")
	invalidBindHostPath        = errortype("invalid binds hostpath, must be absolute path")
	invalidBindContainerPath   = errortype("invalid binds containerpath, must be absolute path")
	invalidBindType            = errortype("invalid bind type")
	invalidPort                = errortype("invalid port defined in yaml")
	invalidExpose              = errortype("invalid expose defined in yaml")
	invalidVolume              = errortype("invalid volume defined in yaml")
	invalidAddonPlan           = errortype("invalid addon plan in yaml")
	invalidImage               = errortype("invalid image defined in yaml")
	invalidTrafficSecurityMode = errortype("invalid traffic security mode in yaml, must be 'https'")
	emptyEndpointDomain        = errortype("empty domain in endpoints")
	invalidEndpointDomain      = errortype("invalid domain in endpoints")
	invalidEndpointPath        = errortype("invalid path in endpoints, must start with '/'")
)

type errortype string

func (e errortype) Error() string {
	return string(e)
}
