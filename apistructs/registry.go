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

// RegistryManifestsRemoveRequest 删除指定集群Registry镜像元数据请求
// POST /api/clusters/{idOrName}/registry/manifests/actions/remove
type RegistryManifestsRemoveRequest struct {
	Images      []string `json:"images"`      // 待删除元数据的镜像列表
	RegistryURL string   `json:"registryURL"` // Registry地址, 接口自动根据集群配置赋值
}

// RegistryManifestsRemoveResponse 删除指定集群Registry镜像元数据响应
type RegistryManifestsRemoveResponse struct {
	Header
	Data RegistryManifestsRemoveResponseData `json:"data"`
}

// RegistryManifestsRemoveResponseData 删除指定集群Registry镜像元数据成功和失败信息
type RegistryManifestsRemoveResponseData struct {
	// 删除元数据成功的镜像列表
	Succeed []string `json:"succeed"`

	// 删除元数据失败的镜像列表和失败原因
	Failed map[string]string `json:"failed"`
}

// RegistryReadonlyResponse 查询指定集群Registry是否只读状态响应
// GET /api/clusters/{idOrName}/registry/readonly
type RegistryReadonlyResponse struct {
	Header
	Data bool `json:"data"` // true只读, false读写
}

//RegistryAuthJson dockerRegistry的认证串
type RegistryUserInfo struct {
	Auth string `json:"auth"`
}
type RegistryAuthJson struct {
	Auths map[string]RegistryUserInfo `json:"auths"`
}
