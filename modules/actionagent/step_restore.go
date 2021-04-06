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
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/actionagent/agenttool"
	"github.com/erda-project/erda/modules/pipeline/pipengine/pvolumes"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/httpclient"
)

func (agent *Agent) restore() {
	for _, in := range agent.Arg.Context.InStorages {
		switch in.Type {

		case string(spec.StoreTypeNFS):
			tarFile := strings.TrimPrefix(in.Value, spec.StoreTypeNFSProto)
			tarDir := agent.EasyUse.ContainerContext
			err := agenttool.UnTar(tarFile, tarDir)
			if err != nil {
				if in.Optional {
					logrus.Printf("[restore] ignore optional restore, type: %s, (prepare to untar [%s] into [%s]).\n",
						spec.StoreTypeNFS, tarFile, tarDir)
					continue
				}
				agent.AppendError(err)
			}
		case string(spec.StoreTypeDiceVolumeLocal), string(spec.StoreTypeDiceVolumeFake):
			// nothing
		case string(spec.StoreTypeOSS):
			agent.AppendError(errors.New("[restore] not implemented now for OSS storage"))

		// dice-nfs-volume 类型，restore 时将 volume.path 下的 data (.tar) 解压到 containerContext 下
		case string(spec.StoreTypeDiceVolumeNFS):
			tarFile := filepath.Join(in.Value, "data")
			tarDir := agent.EasyUse.ContainerContext
			err := agenttool.UnTar(tarFile, tarDir)
			if err != nil {
				if in.Optional {
					logrus.Printf("[restore] ignore optional restore, type: %s, (prepare to untar [%s] into [%s]).\n",
						spec.StoreTypeDiceVolumeNFS, tarFile, tarDir)
					continue
				}
				agent.AppendError(err)
			}
		case string(spec.StoreTypeDiceCacheNFS):
			tarExecDir := filepath.Dir(in.Labels[pvolumes.TaskCachePath])
			tarFile := in.Value + "/" + in.Labels[pvolumes.TaskCacheHashName] + pvolumes.TaskCacheCompressionSuffix
			if filehelper.CheckExist(tarFile, false) != nil {
				logrus.Printf("not get action cache: %s", in.Labels[pvolumes.TaskCachePath])
				continue
			}
			if err := agenttool.UnTar(tarFile, tarExecDir); err != nil {
				logrus.Printf("StoreTypeDiceCacheNFS untar error: %v", err)
			}
			logrus.Printf("get action cache: %s success", in.Labels[pvolumes.TaskCachePath])
		default:
			agent.AppendError(errors.Errorf("[restore] unsupported store type: %s", in.Type))
		}
	}

	for _, f := range agent.Arg.Context.CmsDiceFiles {
		// invoke openapi /api/files?file=${uuid} to download files
		respBody, resp, err := httpclient.New(httpclient.WithCompleteRedirect()).
			Get(agent.EasyUse.OpenAPIAddr).
			Path("/api/files").
			Param("file", f.Labels[pvolumes.VoLabelKeyDiceFileUUID]).
			Header("Authorization", agent.EasyUse.TokenForBootstrap).
			Do().StreamBody()
		if err != nil {
			agent.AppendError(errors.Errorf("failed to download cms file, uuid: %s, err: %v",
				f.Labels[pvolumes.VoLabelKeyDiceFileUUID], err))
			continue
		}
		if !resp.IsOK() {
			bodyBytes, _ := ioutil.ReadAll(respBody)
			agent.AppendError(errors.Errorf("failed to download cms file, uuid: %s, err: %v",
				f.Labels[pvolumes.VoLabelKeyDiceFileUUID], string(bodyBytes)))
			continue
		}
		if err := filehelper.CreateFile2(f.Value, respBody, 0755); err != nil {
			agent.AppendError(err)
			continue
		}
	}
}
