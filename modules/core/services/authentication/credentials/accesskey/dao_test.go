// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package accesskey

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var (
	_mockErr       = fmt.Errorf("mock error")
	_mockTime      = time.Date(2021, 8, 18, 0, 0, 0, 0, time.UTC)
	_mockAccessKey = AccessKey{
		ID:          "aaa",
		AccessKey:   "xxx",
		SecretKey:   "yyy",
		Status:      pb.StatusEnum_ACTIVATE,
		SubjectType: pb.SubjectTypeEnum_MICRO_SERVICE,
		Subject:     "1",
		Description: "xxx",
		CreatedAt:   _mockTime,
		UpdatedAt:   _mockTime,
	}
)

type mockDao struct {
	errorTrigger bool
}

func (m *mockDao) QueryAccessKey(ctx context.Context, req *pb.QueryAccessKeysRequest) ([]AccessKey, int64, error) {
	if m.errorTrigger {
		return nil, 0, _mockErr
	}
	return []AccessKey{
		_mockAccessKey,
	}, 0, nil
}

func (m *mockDao) CreateAccessKey(ctx context.Context, req *pb.CreateAccessKeyRequest) (*AccessKey, error) {
	if m.errorTrigger {
		return nil, _mockErr
	}
	return &_mockAccessKey, nil
}

func (m *mockDao) GetAccessKey(ctx context.Context, req *pb.GetAccessKeyRequest) (*AccessKey, error) {
	if m.errorTrigger {
		return nil, _mockErr
	}
	return &_mockAccessKey, nil
}

func (m *mockDao) UpdateAccessKey(ctx context.Context, req *pb.UpdateAccessKeyRequest) error {
	if m.errorTrigger {
		return _mockErr
	}
	return nil
}

func (m *mockDao) DeleteAccessKey(ctx context.Context, req *pb.DeleteAccessKeyRequest) error {
	if m.errorTrigger {
		return _mockErr
	}
	return nil
}

func newMockDB() *gorm.DB {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	// construct
	rows := sqlmock.NewRows([]string{"id", "access_key", "secret_key", "status", "subject_type", "subject", "description"}).
		AddRow("aaa", "abc", "edf", 1, 1, "1", "xx")
	mock.ExpectQuery("SELECT \\* FROM `erda_access_key`.*?").WillReturnRows(rows)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `erda_access_key`.*?").WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	gdb, err := gorm.Open("mysql", sqlDB)
	if err != nil {
		panic(err)
	}
	return gdb
}

func Test_dao_QueryAccessKey(t *testing.T) {
	d := dao{db: newMockDB()}

	obj, cnt, err := d.QueryAccessKey(context.TODO(), &pb.QueryAccessKeysRequest{
		Status:      pb.StatusEnum_DISABLED,
		SubjectType: pb.SubjectTypeEnum_MICRO_SERVICE,
		AccessKey:   "abc",
		Subject:     "1",
		PageNo:      1,
		PageSize:    2,
	})
	ass := assert.New(t)
	ass.Equal("aaa", obj[0].ID)
	ass.Equal(int64(1), cnt)
	ass.Nil(err)
}
