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

package accesskey

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
	"github.com/erda-project/erda/pkg/secret"
	"github.com/jinzhu/gorm"
)

type dao struct {
	db *gorm.DB
}

func (d *dao) QueryAccessKey(ctx context.Context, req *pb.QueryAccessKeysRequest) ([]AccessKey, error) {
	var objs []AccessKey
	q := d.db
	where := make(map[string]interface{})
	if req.Status != pb.StatusEnum_NOT_SPECIFIED {
		where["status"] = req.Status
	} else {
		q = q.Where("status != ?", pb.StatusEnum_DELETED)
	}
	if req.Subject != "" {
		where["subject"] = req.Subject
	}
	if req.SubjectType != pb.SubjectTypeEnum_NOT_SPECIFIED {
		where["subject_type"] = req.SubjectType
	}
	if req.AccessKey != "" {
		where["access_key"] = req.AccessKey
	}

	if res := q.Where(where).Find(&objs); res.Error != nil {
		return nil, res.Error
	}
	return objs, nil
}

func (d *dao) CreateAccessKey(ctx context.Context, req *pb.CreateAccessKeysRequest) (*AccessKey, error) {
	obj := toModel(req)
	q := d.db.Create(&obj)
	if q.Error != nil {
		return nil, q.Error
	}
	return &obj, nil
}

func toModel(req *pb.CreateAccessKeysRequest) AccessKey {
	pair := secret.CreateAkSkPair()
	return AccessKey{
		AccessKey:   pair.AccessKeyID,
		SecretKey:   pair.SecretKey,
		Status:      pb.StatusEnum_ACTIVATE,
		SubjectType: req.SubjectType,
		Subject:     req.Subject,
		Description: req.Description,
	}
}
