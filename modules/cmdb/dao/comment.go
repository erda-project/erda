package dao

import (
	"github.com/erda-project/erda/modules/cmdb/model"
)

// CreateComment 创建工单评论
func (client *DBClient) CreateComment(comment *model.Comment) error {
	return client.Create(comment).Error
}

// UpdateComment 更新工单评论
func (client *DBClient) UpdateComment(comment *model.Comment) error {
	return client.Save(comment).Error
}

// GetCommentByID 根据commentID获取评论
func (client *DBClient) GetCommentByID(commentID int64) (*model.Comment, error) {
	var comment model.Comment
	if err := client.Where("id = ?", commentID).Find(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetCommentsByTicketID 根据ticketID获取工单评论
func (client *DBClient) GetCommentsByTicketID(ticketID int64, pageNo, pageSize int) (int64, []model.Comment, error) {
	var (
		total    int64
		comments []model.Comment
	)
	if err := client.Where("ticket_id = ?", ticketID).
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&comments).Error; err != nil {
		return 0, nil, err
	}
	if err := client.Model(&model.Comment{}).Where("ticket_id = ?", ticketID).Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, comments, nil
}

// GetLastCommentByTicket 根据 ticketID 获取最新评论
func (client *DBClient) GetLastCommentByTicket(ticketID int64) (*model.Comment, error) {
	var comment model.Comment
	if err := client.Where("ticket_id = ?", ticketID).Order("created_at DESC").First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}
