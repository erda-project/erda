package dao

import (
	"github.com/erda-project/erda/modules/cmdb/model"
)

// CreateCurrentOrg 添加用户当前所属企业
func (client *DBClient) CreateCurrentOrg(orgRelation *model.CurrentOrg) error {
	return client.Create(orgRelation).Error
}

// UpdateCurrentOrg 更新用户当前所属企业
func (client *DBClient) UpdateCurrentOrg(userID string, orgID int64) error {
	var currentOrg model.CurrentOrg
	if err := client.Where("user_id = ?", userID).Find(&currentOrg).Error; err != nil {
		return err
	}
	currentOrg.OrgID = orgID
	return client.Save(&currentOrg).Error
}

// DeleteCurrentOrg 删除当前用户所属企业
func (client *DBClient) DeleteCurrentOrg(userID string) error {
	return client.Where("user_id = ?", userID).Delete(&model.CurrentOrg{}).Error
}

// GetCurrentOrgByUser 根据userID获取当前所属企业ID
func (client *DBClient) GetCurrentOrgByUser(userID string) (int64, error) {
	var currentOrg model.CurrentOrg
	if err := client.Where("user_id = ?", userID).Find(&currentOrg).Error; err != nil {
		return 0, err
	}
	return currentOrg.OrgID, nil
}
