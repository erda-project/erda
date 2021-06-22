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

package bannerImage

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type ComponentImage struct {
	ctxBdl  protocol.ContextBundle
	Version string `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Props   Props  `json:"props,omitempty"`
	State   State  `json:"state,omitempty"`
}

type Props struct {
	Visible bool   `json:"visible"`
	Size    string `json:"size"`
	Src     string `json:"src"`
}

type State struct {
	IsEmpty bool `json:"isEmpty"`
}
