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

// Package registry docker registry manifest操作
package registry

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
)

func DeleteManifests(clusterName string, images []string) error {
	if len(images) == 0 {
		return nil
	}

	imageReq := &apistructs.RegistryManifestsRemoveRequest{
		Images: images,
	}

	var imageResp apistructs.RegistryManifestsRemoveResponse
	path := fmt.Sprintf("/api/clusters/%s/registry/manifests/actions/remove", clusterName)
	resp, err := httpclient.New().Post(discover.Ops()).
		Path(path).
		Header("Content-Type", "application/json").
		JSONBody(imageReq).
		Do().
		JSON(&imageResp)
	if err != nil {
		return errors.Errorf("recycle image: %+v error: %v", images, err)
	}
	if !resp.IsOK() || !imageResp.Success {
		return errors.Errorf("recycle image: %+v fail, statusCode: %d, err: %+v", images, resp.StatusCode(), imageResp.Error)
	}
	if len(imageResp.Data.Failed) > 0 {
		return errors.Errorf("recycle image fail: %+v", imageResp.Data.Failed)
	}

	return nil
}
