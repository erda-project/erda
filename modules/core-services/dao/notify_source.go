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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
)

func (client *DBClient) GetNotifySourcesByNotifyID(notifyID int64) ([]*apistructs.NotifySource, error) {
	var items []*apistructs.NotifySource
	err := client.Table("dice_notify_sources").
		Joins("inner join dice_notifies on dice_notifies.id = dice_notify_sources.notify_id").
		Select("dice_notify_sources.id, dice_notify_sources.name,dice_notify_sources.source_type,dice_notify_sources.source_id").
		Where("dice_notify_sources.notify_id = ?", notifyID).
		Scan(&items).Error
	return items, err
}

func (client *DBClient) DeleteNotifySource(request *apistructs.DeleteNotifySourceRequest) error {
	err := client.Where("source_type =? and source_id= ? and org_id = ?",
		request.SourceType, request.SourceID, request.OrgID).
		Delete(&model.NotifySource{}).Error
	return err
}
