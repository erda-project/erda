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
	"time"

	"github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
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

func (m *mockDao) QueryAccessKey(ctx context.Context, req *pb.QueryAccessKeysRequest) ([]AccessKey, error) {
	if m.errorTrigger {
		return nil, _mockErr
	}
	return []AccessKey{
		_mockAccessKey,
	}, nil
}

func (m *mockDao) CreateAccessKey(ctx context.Context, req *pb.CreateAccessKeysRequest) (*AccessKey, error) {
	if m.errorTrigger {
		return nil, _mockErr
	}
	return &_mockAccessKey, nil
}

func (m *mockDao) GetAccessKey(ctx context.Context, req *pb.GetAccessKeysRequest) (*AccessKey, error) {
	if m.errorTrigger {
		return nil, _mockErr
	}
	return &_mockAccessKey, nil
}

func (m *mockDao) UpdateAccessKey(ctx context.Context, req *pb.UpdateAccessKeysRequest) error {
	if m.errorTrigger {
		return _mockErr
	}
	return nil
}

func (m *mockDao) DeleteAccessKey(ctx context.Context, req *pb.DeleteAccessKeysRequest) error {
	if m.errorTrigger {
		return _mockErr
	}
	return nil
}
