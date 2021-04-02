package dao

import (
	"context"

	"github.com/erda-project/erda/modules/cmdb/types"

	"github.com/pkg/errors"
)

// CreateOrUpdateService 更新服务信息
func (client *DBClient) CreateOrUpdateService(ctx context.Context, service *types.CmService) error {
	var err error

	if service == nil {
		return errors.Errorf("invalid params: service is nil")
	}

	if err = client.Save(service).Error; err != nil {
		return err
	}

	return nil
}
