package dbclient

import (
	"github.com/erda-project/erda/modules/pipeline/spec"

	"github.com/pkg/errors"
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
