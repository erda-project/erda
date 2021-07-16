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

package dao

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
)

func (client *DBClient) CreateMBox(mbox *model.MBox) error {
	return client.Create(&mbox).Error
}

// UpdateMbox update mbox
func (client *DBClient) UpdateMbox(mbox *model.MBox) error {
	return client.Save(mbox).Error
}

func (client *DBClient) DeleteMBox(id int64) error {
	var mbox model.MBox
	err := client.Where("id = ?", id).Find(&mbox).Error
	if err != nil {
		return err
	}
	return client.Delete(&mbox).Error
}

func (client *DBClient) SetMBoxReadStatus(request *apistructs.SetMBoxReadStatusRequest) error {
	now := time.Now()
	return client.Model(&model.MBox{}).
		Where("status =?", apistructs.MBoxUnReadStatus).
		Where("id in (?) and org_id=? and user_id =?", request.IDs, request.OrgID, request.UserID).
		Updates(map[string]interface{}{"read_at": &now, "status": apistructs.MBoxReadStatus,
			"unread_count": 0}).Error
}

func (client *DBClient) GetMBoxStats(orgID int64, userID string) (*apistructs.QueryMBoxStatsData, error) {
	var count, unReadCount int64
	if err := client.Model(&model.MBox{}).Where("org_id=? and user_id =?", orgID, userID).
		Where("status =? ", apistructs.MBoxUnReadStatus).Count(&count).Error; err != nil {
		return nil, err
	}
	var mbox []model.MBox
	if err := client.Model(&model.MBox{}).Where("org_id=? and user_id =?", orgID, userID).
		Where("status =? ", apistructs.MBoxUnReadStatus).Where("unread_count > 1 and deduplicate_id != \"\"").
		Find(&mbox).Count(&unReadCount).Error; err != nil {
		return nil, err
	}
	count = count - unReadCount
	for _, v := range mbox {
		count = count + v.UnreadCount
	}
	return &apistructs.QueryMBoxStatsData{
		UnreadCount: count,
	}, nil
}

func (client *DBClient) GetMBox(id int64, orgID int64, userID string) (*apistructs.MBox, error) {
	var mbox model.MBox
	err := client.Where("id =? and org_id=? and user_id =?", id, orgID, userID).First(&mbox).Error
	if err != nil {
		return nil, err
	}
	return mbox.ToApiData(), nil
}

func (client *DBClient) QueryMBox(request *apistructs.QueryMBoxRequest) (*apistructs.QueryMBoxData, error) {
	var mboxList []model.MBox
	query := client.Model(&model.MBox{}).Where("org_id =? and user_id= ?", request.OrgID, request.UserID)
	if request.Status != "" {
		query = query.Where("status =? ", request.Status)
	}
	query = query.Order("created_at desc").Select("id,title,status,label,created_at,read_at,deduplicate_id,unread_count")
	var count int
	err := query.Count(&count).Error
	if err != nil {
		return nil, err
	}

	err = query.Offset((request.PageNo - 1) * request.PageSize).
		Limit(request.PageSize).
		Find(&mboxList).Error
	if err != nil {
		return nil, err
	}

	result := &apistructs.QueryMBoxData{
		Total: count,
		List:  []*apistructs.MBox{},
	}
	for _, mboxItem := range mboxList {
		result.List = append(result.List, mboxItem.ToApiData())
	}
	return result, nil
}

// GetMboxByDeduplicateID get a mbox by deduplicate_id
func (client *DBClient) GetMboxByDeduplicateID(orgID int64, deduplicateID, userID string) (*model.MBox, error) {
	var mbox model.MBox
	if err := client.Where("deduplicate_id = ? and org_id = ? and user_id = ?", deduplicateID, orgID,
		userID).First(&mbox).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &mbox, nil
}
