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

package dbclient

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) GetBuildCache(clusterName, imageName string) (cache spec.CIV3BuildCache, err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to get build cache, clusterName [%s], imageName [%s]", clusterName, imageName)
	}()

	cache.ClusterName = clusterName
	cache.Name = imageName
	ok, err := client.Get(&cache)
	if err != nil {
		return spec.CIV3BuildCache{}, err
	}
	if !ok {
		return spec.CIV3BuildCache{}, errors.New("not found")
	}
	return cache, nil
}

func (client *Client) DeleteBuildCache(id interface{}) (err error) {
	defer func() { err = errors.Wrapf(err, "failed to delete build cache, id [%v]", id) }()

	_, err = client.ID(id).Delete(&spec.CIV3BuildCache{})
	return err
}
