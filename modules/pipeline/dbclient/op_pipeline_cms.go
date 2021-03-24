package dbclient

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) IdempotentCreateCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...SessionOption) (spec.PipelineCmsNs, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	// get cmsNs by source + ns
	cmsNs, exist, err := client.GetCmsNs(pipelineSource, ns, ops...)
	if err != nil {
		return spec.PipelineCmsNs{}, err
	}
	if exist {
		return cmsNs, nil
	}

	// not exist, create
	var newNS spec.PipelineCmsNs
	newNS.PipelineSource = pipelineSource
	newNS.Ns = ns
	_, err = session.InsertOne(&newNS)
	return newNS, err
}

func (client *Client) IdempotentDeleteCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...SessionOption) error {
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
	_, err = session.ID(cmsNs.ID).Delete(&spec.PipelineCmsNs{})
	return err
}

func (client *Client) GetCmsNs(pipelineSource apistructs.PipelineSource, ns string, ops ...SessionOption) (spec.PipelineCmsNs, bool, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var cmsNs spec.PipelineCmsNs
	cmsNs.PipelineSource = pipelineSource
	cmsNs.Ns = ns
	exist, err := session.Get(&cmsNs)
	if err != nil {
		return spec.PipelineCmsNs{}, false, err
	}
	return cmsNs, exist, nil
}

func (client *Client) PrefixListNs(pipelineSource apistructs.PipelineSource, nsPrefix string, ops ...SessionOption) ([]spec.PipelineCmsNs, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var namespaces []spec.PipelineCmsNs
	if err := session.Where("ns LIKE ?", nsPrefix+"%").Find(&namespaces); err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (client *Client) UpdateCmsNsConfigs(cmsNs spec.PipelineCmsNs, configs []spec.PipelineCmsConfig, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	for _, config := range configs {
		if err := client.InsertOrUpdateCmsNsConfig(cmsNs, config, ops...); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) InsertOrUpdateCmsNsConfig(cmsNs spec.PipelineCmsNs, config spec.PipelineCmsConfig, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	var query spec.PipelineCmsConfig
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

func (client *Client) DeleteCmsNsConfigs(cmsNs spec.PipelineCmsNs, keys []string, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	if len(keys) == 0 {
		return nil
	}

	_, err := session.In("key", keys).Delete(&spec.PipelineCmsConfig{NsID: cmsNs.ID})
	return err
}

func (client *Client) GetCmsNsConfigs(cmsNs spec.PipelineCmsNs, keys []string, ops ...SessionOption) ([]spec.PipelineCmsConfig, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var configs []spec.PipelineCmsConfig
	if len(keys) > 0 {
		session.In("key", keys)
	}
	err := session.Find(&configs, &spec.PipelineCmsConfig{NsID: cmsNs.ID})
	if err != nil {
		return nil, err
	}
	return configs, nil
}
