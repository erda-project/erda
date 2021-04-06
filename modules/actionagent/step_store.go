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

package actionagent

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/actionagent/agenttool"
	"github.com/erda-project/erda/modules/pipeline/pipengine/pvolumes"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/filehelper"
)

func (agent *Agent) store() {
	for _, out := range agent.Arg.Context.OutStorages {
		switch out.Type {

		case string(spec.StoreTypeNFS):
			tarFile := strings.TrimPrefix(out.Value, spec.StoreTypeNFSProto)
			tarDir := filepath.Join(agent.EasyUse.ContainerContext, out.Name)
			err := agenttool.Tar(tarFile, tarDir)
			if err != nil {
				agent.AppendError(err)
			}
		case string(spec.StoreTypeDiceVolumeLocal), string(spec.StoreTypeDiceVolumeFake):
			// nothing
		case string(spec.StoreTypeOSS):
			agent.AppendError(errors.New("[store] not implemented now for OSS storage"))

		// dice-nfs-volume 类型，store 时将对应 containerContext 下的 task namespace 整个压缩为 volume.path 下的 data (.tar)
		case string(spec.StoreTypeDiceVolumeNFS):
			tarFile := filepath.Join(out.Value, "data")
			tarDir := filepath.Join(agent.EasyUse.ContainerContext, out.Name)
			err := agenttool.Tar(tarFile, tarDir)
			if err != nil {
				agent.AppendError(err)
			}
		case string(spec.StoreTypeDiceCacheNFS):
			tarFile := out.Value + "/" + out.Labels[pvolumes.TaskCacheHashName] + pvolumes.TaskCacheCompressionSuffix
			if filehelper.CheckExist(out.Labels[pvolumes.TaskCachePath], true) != nil {
				logrus.Printf("upload action cache error: %s is not dir", out.Labels[pvolumes.TaskCachePath])
				continue
			}
			if err := agenttool.Tar(tarFile, out.Labels[pvolumes.TaskCachePath]); err != nil {
				logrus.Printf("StoreTypeDiceCacheNFS tar error: %v", err)
			}
			logrus.Printf("upload action cache %s success", out.Labels[pvolumes.TaskCachePath])
		default:
			agent.AppendError(errors.Errorf("[store] unsupported store type: %s", out.Type))
		}
	}
}
