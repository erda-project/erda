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

package pipelineyml

type Source map[string]interface{}

type Params map[string]interface{}

type Version map[string]string

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

const (
	RES_TYPE_BUILDPACK = "buildpack"
	RES_TYPE_BP_IMAGE  = "bp-image"
	RES_TYPE_SONAR     = "sonar"
	RES_TYPE_UT        = "ut"

	RES_TYPE_GIT = "git"
)
