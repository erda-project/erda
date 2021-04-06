package model

import "github.com/erda-project/erda/apistructs"

// 应用和发布项关联关系
type ApplicationPublishItemRelation struct {
	BaseModel
	AppID         int64
	PublishItemID int64
	Env           apistructs.DiceWorkspace
	Creator       string
	AK            string
	AI            string
}

// TableName 设置模型对应数据库表名称
func (ApplicationPublishItemRelation) TableName() string {
	return "dice_app_publish_item_relation"
}
