package dao

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

// Notice 平台公告
type Notice struct {
	BaseModel
	OrgID   uint64
	Content string
	Status  apistructs.NoticeStatus
	Creator string
}

// TableName 表名
func (Notice) TableName() string {
	return "dice_notices"
}

// CreateNotice 创建公告
func (client *DBClient) CreateNotice(notice *Notice) error {
	return client.Create(notice).Error
}

// UpdateNotice 编辑公告
func (client *DBClient) UpdateNotice(notice *Notice) error {
	return client.Save(notice).Error
}

// DeleteNotice 删除公告
func (client *DBClient) DeleteNotice(noticeID uint64) error {
	return client.Where("id = ?", noticeID).Delete(Notice{}).Error
}

func (client *DBClient) GetNotice(noticeID uint64) (*Notice, error) {
	var notice Notice
	if err := client.Where("id = ?", noticeID).Find(&notice).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &notice, nil
}

// ListNotice 公告列表
func (client *DBClient) ListNotice(req *apistructs.NoticeListRequest) (uint64, []Notice, error) {
	var (
		total   uint64
		notices []Notice
	)
	cond := Notice{}
	if req.OrgID > 0 {
		cond.OrgID = req.OrgID
	}
	if req.Status != "" {
		cond.Status = req.Status
	}
	sql := client.Where(cond)
	if req.Content != "" {
		sql = sql.Where("content LIKE ?", "%"+req.Content+"%")
	}
	if err := sql.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).
		Find(&notices).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, notices, nil
}
