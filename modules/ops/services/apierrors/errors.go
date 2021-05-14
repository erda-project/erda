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

package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var (
	ErrCreateFoobar = err("ErrCreateFoobar", "failed to create foobar")
)

var (
	ErrGetAddonConfig    = err("ErrGetAddonConfig", "failed to get addon configuration")
	ErrGetAddonStatus    = err("ErrGetAddonStatus", "failed to  get addon status")
	ErrUpdateAddonConfig = err("ErrUpdateAddonConfig", "failed to update addon configuration")
	ErrCreateCluster     = err("ErrCreateCluster", "failed to create cluster")
)

var (
	ErrListEdgeSite         = err("ErrListEdgeSite", "failed to list edge site")
	ErrGetEdgeSite          = err("ErrGetEdgeSite", "failed to get edge site")
	ErrCreateEdgeSite       = err("ErrCreateEdgeSite", "failed to create edge site")
	ErrUpdateEdgeSite       = err("ErrUpdateEdgeSite", "failed to update edge site")
	ErrDeleteEdgeSite       = err("ErrDeleteEdgeSite", "failed to delete edge site")
	ErrGetEdgeSiteInit      = err("ErrGetEdgeSiteInit", "failed to get edge site init shell")
	ErrOfflineEdgeSite      = err("ErrOfflineEdgeSite", "failed to offline edge site")
	ErrListEdgeConfigSet    = err("ErrListEdgeConfigSet", "failed to list configSet")
	ErrCreateEdgeConfigSet  = err("ErrCreateEdgeConfigSet", "failed to create edge  configSet")
	ErrUpdateEdgeConfigSet  = err("ErrUpdateEdgeConfigSet", "failed to update edge  configSet")
	ErrDeleteEdgeConfigSet  = err("ErrDeleteEdgeConfigSet", "failed to delete edge  configSet")
	ErrListEdgeCfgSetItem   = err("ErrListEdgeCfgSetItem", "failed to list  configSet")
	ErrCreateEdgeCfgSetItem = err("ErrCreateEdgeCfgSetItem", "failed to create edge configSet item")
	ErrUpdateEdgeCfgSetItem = err("ErrUpdateEdgeCfgSetItem", "failed to update edge configSet item")
	ErrDeleteEdgeCfgSetItem = err("ErrDeleteEdgeCfgSetItem", "failed to delete edge configSet item")
	ErrListEdgeApp          = err("ErrListEdgeSite", "failed to list edge app")
	ErrCreateEdgeApp        = err("ErrCreateEdgeApp", "failed to create edge app")
	ErrUpdateEdgeApp        = err("ErrUpdateEdgeSite", "failed to update edge app")
	ErrDeleteEdgeApp        = err("ErrDeleteEdgeSite", "failed to delete edge app")
	ErrRestartEdgeApp       = err("ErrRestartEdgeApp", "failed to restart edge app")
	ErrOfflineEdgeAppSite   = err("ErrOfflineEdgeAppSite", "failed to offline specified site in edge app ")
	AccessDeny              = err("ErrAccessDeny", "permission denied")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
