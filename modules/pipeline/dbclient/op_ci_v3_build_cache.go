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
