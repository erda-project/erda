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

import (
	"net/url"
	"path/filepath"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// ParseDiceYmlJobBinds 将 diceYmlJob 里老格式的 binds 转换为新的格式
func ParseDiceYmlJobBinds(diceYmlJob *diceyml.Job) ([]apistructs.Bind, error) {
	binds, err := diceyml.ParseBinds(diceYmlJob.Binds)
	if err != nil {
		return nil, err
	}
	var result []apistructs.Bind
	for _, bind := range binds {
		result = append(result, apistructs.Bind{
			ContainerPath: bind.ContainerPath,
			HostPath:      bind.HostPath,
			ReadOnly:      bind.Type == "r",
		})
	}
	return result, nil
}

// GenerateTaskCommonBinds 生成 task 通用 binds
func GenerateTaskCommonBinds(mountPoint string) []apistructs.Bind {

	const (
		dockerSock = "/var/run/docker.sock"
	)

	var binds []apistructs.Bind
	dockerSockBind := apistructs.Bind{
		HostPath:      dockerSock,
		ContainerPath: dockerSock,
		ReadOnly:      true,
	}
	binds = append(binds, dockerSockBind)

	storageURL := conf.StorageURL()
	URL, _ := url.Parse(storageURL)
	if URL.Scheme == "file" {
		var storageBind apistructs.Bind
		_path := filepath.Join(mountPoint, URL.Path)
		storageBind = apistructs.Bind{
			HostPath:      _path,
			ContainerPath: _path,
			ReadOnly:      false,
		}
		binds = append(binds, storageBind)
	}
	return binds
}
