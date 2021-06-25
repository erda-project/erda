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
	registryUrl := clusterInfo.MustGet(apistructs.REGISTRY_ADDR)
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
