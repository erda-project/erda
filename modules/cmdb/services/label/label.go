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

package label

import (
	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
)

// Label 标签封装
type Label struct {
	db *dao.DBClient
}

// Option 定义 Label 对象的配置选项
type Option func(*Label)

// New 新建 Label 实例
func New(options ...Option) *Label {
	label := &Label{}
	for _, op := range options {
		op(label)
	}
	return label
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(l *Label) {
		l.db = db
	}
}

// Create 创建标签
func (l *Label) Create(req *apistructs.ProjectLabelCreateRequest) (int64, error) {
	// 参数校验
	if err := l.checkCreateParam(req); err != nil {
		return 0, apierrors.ErrCreateLabel.InvalidParameter(err)
	}

	// 同名检查
	old, err := l.db.GetLabelByName(req.ProjectID, req.Name)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return 0, err
	}
	if old != nil {
		return 0, apierrors.ErrCreateLabel.AlreadyExists()
	}

	// 标签信息入库
	label := &dao.Label{
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		ProjectID: req.ProjectID,
		Creator:   req.UserID,
	}
	if err := l.db.CreateLabel(label); err != nil {
		return 0, err
	}

	return label.ID, nil
}

// Delete 删除工单
func (l *Label) Delete(labelID int64) error {
	l.db.DeleteLabelRelationsByLabel(uint64(labelID))
	return l.db.DeleteLabel(labelID)
}

// Update 更新标签
func (l *Label) Update(req *apistructs.ProjectLabelUpdateRequest) error {
	label, err := l.db.GetLabel(req.ID)
	if err != nil {
		return err
	}
	if label == nil {
		return apierrors.ErrUpdateLabel.NotFound()
	}

	// 新名称 label 是否存在
	new, err := l.db.GetLabelByName(label.ProjectID, req.Name)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	if new != nil && new.Name != label.Name {
		return apierrors.ErrUpdateLabel.AlreadyExists()
	}

	label.Name = req.Name
	label.Color = req.Color

	return l.db.UpdateLabel(label)
}

// List 标签列表
func (l *Label) List(req *apistructs.ProjectLabelListRequest) (int64, []apistructs.ProjectLabel, error) {
	if req.ProjectID == 0 {
		return 0, nil, apierrors.ErrGetLabels.MissingParameter("projectID")
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	total, labels, err := l.db.ListLabel(req)
	if err != nil {
		return 0, nil, apierrors.ErrGetLabels.InternalError(err)
	}
	labelResp := make([]apistructs.ProjectLabel, 0, len(labels))
	for _, v := range labels {
		labelResp = append(labelResp, l.convert(v))
	}

	return total, labelResp, nil
}

// GetByID 标签ByID
func (l *Label) GetByID(id int64) (apistructs.ProjectLabel, error) {
	label, err := l.db.GetLabel(id)
	if err != nil {
		return apistructs.ProjectLabel{}, err
	}

	return l.convert(*label), nil
}

// ListByNames 根据标签名称获取标签列表
func (l *Label) ListByNames(projectID uint64, names []string) ([]apistructs.ProjectLabel, error) {
	labels, err := l.db.ListLabelByNames(projectID, names)
	if err != nil {
		return nil, err
	}

	labelResp := make([]apistructs.ProjectLabel, 0, len(labels))
	for _, v := range labels {
		labelResp = append(labelResp, l.convert(v))
	}

	return labelResp, nil
}

// CreateRelation 创建标签关联关系
func (l *Label) CreateRelation(lr *dao.LabelRelation) error {
	return l.db.CreateLabelRelation(lr)
}

func (l *Label) DeleteRelations(refType apistructs.ProjectLabelType, refID uint64) error {
	return l.db.DeleteLabelRelations(refType, refID)
}

func (l *Label) checkCreateParam(req *apistructs.ProjectLabelCreateRequest) error {
	if req.Name == "" {
		return errors.Errorf("name is empty")
	}
	if utf8.RuneCountInString(req.Name) > 50 {
		return errors.Errorf("max name length is 50")
	}

	if req.Type == "" {
		return errors.Errorf("type is empty")
	}

	if req.Color == "" {
		return errors.Errorf("color is empty")
	}

	if req.ProjectID == 0 {
		return errors.Errorf("projectID is empty")
	}

	return nil
}

func (l *Label) convert(label dao.Label) apistructs.ProjectLabel {
	return apistructs.ProjectLabel{
		ID:        label.ID,
		Name:      label.Name,
		Type:      label.Type,
		Color:     label.Color,
		ProjectID: label.ProjectID,
		Creator:   label.Creator,
		CreatedAt: label.CreatedAt,
		UpdatedAt: label.UpdatedAt,
	}
}
