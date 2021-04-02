package model

import "github.com/erda-project/erda/apistructs"

type NotifyItem struct {
	BaseModel
	Name             string `gorm:"size:150"`
	DisplayName      string `gorm:"size:150"`
	Category         string `gorm:"size:150"`
	MobileTemplate   string `gorm:"type:text"`
	EmailTemplate    string `gorm:"type:text"`
	DingdingTemplate string `gorm:"type:text"`
	MBoxTemplate     string `gorm:"type:text;column:mbox_template"`
	// 语音通知模版
	VMSTemplate string `gorm:"type:text;column:vms_template"`
	// 语音通知的被叫显号，语音模版属于公共号码池外呼的时候，被叫显号必须是空
	// 属于专属号码外呼的时候，被叫显号不能为空
	CalledShowNumber string `gorm:"size:150;column:called_show_number"`
	ScopeType        string `gorm:"size:150"`
	Label            string `gorm:"size:150"`
	Params           string `gorm:"type:text"`
}

func (NotifyItem) TableName() string {
	return "dice_notify_items"
}

func (notifyItem *NotifyItem) ToApiData() *apistructs.NotifyItem {
	return &apistructs.NotifyItem{
		ID:               notifyItem.ID,
		Name:             notifyItem.Name,
		DisplayName:      notifyItem.DisplayName,
		Category:         notifyItem.Category,
		ScopeType:        notifyItem.ScopeType,
		Label:            notifyItem.Label,
		Params:           notifyItem.Params,
		MobileTemplate:   notifyItem.MobileTemplate,
		VMSTemplate:      notifyItem.VMSTemplate,
		CalledShowNumber: notifyItem.CalledShowNumber,
	}
}
