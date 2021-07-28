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

package db

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
)

var (
	DefaultOperationsForKV        = pb.PipelineCmsConfigOperations{CanDownload: false, CanEdit: true, CanDelete: true}
	DefaultOperationsForDiceFiles = pb.PipelineCmsConfigOperations{CanDownload: true, CanEdit: true, CanDelete: true}
)

var (
	ConfigTypeKV       = "kv"
	ConfigTypeDiceFile = "dice-file"
)

type Client struct {
	mysqlxorm.Interface
}

func (client *Client) IdempotentCreateCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...mysqlxorm.SessionOption) (PipelineCmsNs, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	// get cmsNs by source + ns
	cmsNs, exist, err := client.GetCmsNs(pipelineSource, ns, ops...)
	if err != nil {
		return PipelineCmsNs{}, err
	}
	if exist {
		return cmsNs, nil
	}

	// not exist, create
	var newNS PipelineCmsNs
	newNS.PipelineSource = pipelineSource
	newNS.Ns = ns
	_, err = session.InsertOne(&newNS)
	return newNS, err
}

func (client *Client) IdempotentDeleteCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	// get cmsNs by source + ns
	cmsNs, exist, err := client.GetCmsNs(pipelineSource, ns, ops...)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	// delete ns by id
	if cmsNs.ID == 0 {
		return errors.Errorf("cms ns missing id")
	}
	_, err = session.ID(cmsNs.ID).Delete(&PipelineCmsNs{})
	return err
}

func (client *Client) GetCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...mysqlxorm.SessionOption) (PipelineCmsNs, bool, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var cmsNs PipelineCmsNs
	cmsNs.PipelineSource = pipelineSource
	cmsNs.Ns = ns
	exist, err := session.Get(&cmsNs)
	if err != nil {
		return PipelineCmsNs{}, false, err
	}
	return cmsNs, exist, nil
}

func (client *Client) PrefixListNs(pipelineSource apistructs.PipelineSource, nsPrefix string, ops ...mysqlxorm.SessionOption) ([]PipelineCmsNs, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var namespaces []PipelineCmsNs
	if err := session.Where("ns LIKE ?", nsPrefix+"%").Find(&namespaces); err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (client *Client) UpdateCmsNsConfigs(cmsNs PipelineCmsNs, configs []PipelineCmsConfig, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	for _, config := range configs {
		if err := client.InsertOrUpdateCmsNsConfig(cmsNs, config, ops...); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) InsertOrUpdateCmsNsConfig(cmsNs PipelineCmsNs, config PipelineCmsConfig, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	var query PipelineCmsConfig
	query.NsID = config.NsID
	query.Key = config.Key

	exist, err := session.Get(&query)
	if err != nil {
		return err
	}

	// update
	if exist {
		// no need update
		if query.Equal(config) {
			return nil
		}
		_, err = session.ID(query.ID).Update(&config)
		return err
	}

	// create
	_, err = session.InsertOne(&config)
	return err
}

func (client *Client) DeleteCmsNsConfigs(cmsNs PipelineCmsNs, keys []string, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if len(keys) == 0 {
		return nil
	}

	_, err := session.In("key", keys).Delete(&PipelineCmsConfig{NsID: cmsNs.ID})
	return err
}

func (client *Client) GetCmsNsConfigs(cmsNs PipelineCmsNs, keys []string, ops ...mysqlxorm.SessionOption) ([]PipelineCmsConfig, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var configs []PipelineCmsConfig
	if len(keys) > 0 {
		session.In("key", keys)
	}
	err := session.Find(&configs, &PipelineCmsConfig{NsID: cmsNs.ID})
	if err != nil {
		return nil, err
	}
	return configs, nil
}
