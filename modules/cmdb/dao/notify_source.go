package dao

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
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
