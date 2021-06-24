package dao

import (
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/jinzhu/gorm"
)

// GetHostByClusterAndIP get host info according cluster & privateAddr
func (client *DBClient) GetHostByClusterAndIP(clusterName, privateAddr string) (*model.Host, error) {
	var host model.Host
	if err := client.Where("cluster = ?", clusterName).
		Where("private_addr = ?", privateAddr).First(&host).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &host, nil
}
