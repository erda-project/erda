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

package apistructs

// AddonPrebuildReq addon prebuild request body
type AddonPrebuildReq struct {
	Name       string            `json:"name"`
	Plan       string            `json:"plan"`
	Type       string            `json:"type"`
	InstanceID string            `json:"instanceId"`
	Config     map[string]string `json:"config"`
	Options    map[string]string `json:"options"`
	Actions    map[string]string `json:"actions"`
}

// AddonPrebuildOverlayReq addon prebuild overlay request body，rds覆盖mysql
type AddonPrebuildOverlayReq struct {
	To   string `json:"to"`
	Type string `json:"type"`
}

// SaveAddonPrebuildReq addon prebuild overlay request body，rds覆盖mysql
type SaveAddonPrebuildReq struct {
	Addons        []AddonPrebuildReq        `json:"addons"`
	AddonsOverlay []AddonPrebuildOverlayReq `json:"addons_overlay"`
}
