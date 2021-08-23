// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package registry docker registry manifest操作
package registry

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/registryhelper"
)

func DeleteManifests(bdl *bundle.Bundle, clusterName string, images []string) error {
	if len(images) == 0 {
		return nil
	}
	removeReq := registryhelper.RemoveManifestsRequest{
		Images:     images,
		ClusterKey: clusterName,
	}
	clusterInfo, err := bdl.QueryClusterInfo(clusterName)

	if err != nil {
		return err
	}
	registryUrl := clusterInfo.Get(apistructs.REGISTRY_ADDR)
	if registryUrl == "" {
		return errors.New("registryUrl is empty")
	}
	removeReq.RegistryURL = registryUrl
	removeResp, err := registryhelper.RemoveManifests(removeReq)
	if err != nil {
		return err
	}
	if len(removeResp.Failed) > 0 {
		return errors.Errorf("recycle image fail: %+v", removeResp.Failed)
	}
	return nil
}
