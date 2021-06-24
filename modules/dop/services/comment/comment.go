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

package comment

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/model"
	"github.com/erda-project/erda/modules/pkg/user"
)

// Comment 工单评论操作封装
type Comment struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 Comment 配置选项
type Option func(*Comment)

// New 新建 Comment 实例
func New(options ...Option) *Comment {
	t := &Comment{}
	for _, op := range options {
		op(t)
	}
	return t
}

// WithDBClient 配置 Comment 数据库选项
func WithDBClient(db *dao.DBClient) Option {
	return func(t *Comment) {
		t.db = db
	}
}

// WithBundle 配置 Comment bundle选项
func WithBundle(bdl *bundle.Bundle) Option {
	return func(t *Comment) {
		t.bdl = bdl
	}
}

// Create 创建工单评论
func (c *Comment) Create(userID user.ID, req *apistructs.CommentCreateRequest) (int64, error) {
	if req.UserID != userID.String() {
		return 0, errors.Errorf("user id doesn't match")
	}
	if ticket, err := c.db.GetTicket(req.TicketID); err != nil || ticket == nil {
		return 0, errors.Errorf("invalid ticket id %v, (%v)", req.TicketID, err)
	}

	if req.CommentType == "" {
		req.CommentType = apistructs.NormalTCType
	}
	comment := model.Comment{
		TicketID:    req.TicketID,
		CommentType: req.CommentType,
		Content:     req.Content,
		IRComment:   req.IRComment,
		UserID:      req.UserID,
	}
	if err := c.db.CreateComment(&comment); err != nil {
		return 0, err
	}

	return int64(comment.ID), nil
}

// Update 更新工单评论
func (c *Comment) Update(commentID int64, userID user.ID, req *apistructs.CommentUpdateRequestBody) error {
	if req.Content == "" {
		return errors.Errorf("content is empty")
	}
	comment, err := c.db.GetCommentByID(commentID)
	if err != nil {
		return err
	}
	if comment.UserID != userID.String() {
		return errors.Errorf("user id doesn't match")
	}
	comment.Content = req.Content

	return c.db.UpdateComment(comment)
}

// List 工单评论列表
func (c *Comment) List(ticketID int64, pageNo, pageSize int) (*apistructs.CommentListResponseData, error) {
	total, comments, err := c.db.GetCommentsByTicketID(ticketID, pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	commentDTOs := make([]apistructs.Comment, 0, len(comments))
	for i := range comments {
		commentDTO := c.convertToCommentDTO(&comments[i])
		commentDTOs = append(commentDTOs, *commentDTO)
	}
	return &apistructs.CommentListResponseData{Total: total, Comments: commentDTOs}, nil
}

func (c *Comment) convertToCommentDTO(comment *model.Comment) *apistructs.Comment {
	return &apistructs.Comment{
		CommentID:   int64(comment.ID),
		TicketID:    comment.TicketID,
		CommentType: comment.CommentType,
		Content:     comment.Content,
		IRComment:   comment.IRComment,
		UserID:      comment.UserID,
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
	}
}
