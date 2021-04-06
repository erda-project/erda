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

// DO NOT EDIT!!!
// See go:generate at github.com/erda-project/erda/modules/pipeline/pipelineyml/builtin_resourcetype.go

package pipelineymlv1

var BuiltinResTypeNames = []string{
	string(RES_TYPE_GIT),
	string(RES_TYPE_BUILDPACK),
	string(RES_TYPE_BP_COMPILE),
	string(RES_TYPE_BP_IMAGE),
	string(RES_TYPE_DICE),
	string(RES_TYPE_ABILITY),
	string(RES_TYPE_ADDON_REGISTRY),
	string(RES_TYPE_DICEHUB),
	string(RES_TYPE_IT),
	string(RES_TYPE_SONAR),
	string(RES_TYPE_FLINK),
	string(RES_TYPE_SPARK),
	string(RES_TYPE_UT),
}
