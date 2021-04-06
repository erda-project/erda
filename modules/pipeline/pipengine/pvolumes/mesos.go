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

package pvolumes

import "path"

const (
	EnvKeyMesosFetcherURI = "MESOS_FETCHER_URI"
)

// MakeMesosFetcherURI4AliyunRegistrySecret 生成 DC/OS mesos 下的 fetcherURI，相当于 k8s secret for aliyun docker registry
func MakeMesosFetcherURI4AliyunRegistrySecret(mountPoint string) string {
	return "file://" + path.Clean(mountPoint+"/docker-registry-aliyun/password.tar.gz")
}
