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

package release

type releaseConfig struct {
	MaxTimeReserved string
}

// ResourceType Release type
type ResourceType string

const (
	// ResourceTypeDiceYml ResourceType is dice.yml
	ResourceTypeDiceYml ResourceType = "diceyml"
	// ResourceTypeAddonYml ResourceType is addon.yml
	ResourceTypeAddonYml ResourceType = "addonyml"
	// ResourceTypeBinary ResourceType is binary executable
	ResourceTypeBinary ResourceType = "binary"
	// ResourceTypeScript ResourceType is executable script file, eg: shell/python/ruby, etc
	ResourceTypeScript ResourceType = "script"
	// ResourceTypeSQL ResourceType is sql
	ResourceTypeSQL ResourceType = "sql"
	// ResourceTypeDataSet ResourceType is Data text file
	ResourceTypeDataSet ResourceType = "data"
	// ResourceTypeAndroid ResourceType is android
	ResourceTypeAndroid ResourceType = "android"
	// ResourceTypeIOS ResourceType is ios
	ResourceTypeIOS ResourceType = "ios"
	// ResourceTypeMigration ResourceType is migration文件releaseID
	ResourceTypeMigration ResourceType = "migration"
	// ResourceTypeH5 ResourceType is h5
	ResourceTypeH5 ResourceType = "h5"
)

const (
	// AliYunRegistry Alibaba Cloud registry prefix
	AliYunRegistry = "registry.cn-hangzhou.aliyuncs.com"
)
