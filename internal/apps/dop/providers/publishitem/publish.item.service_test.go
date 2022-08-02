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

package publishitem

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
)

func TestQueryPublishItem(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "QueryPublishItem", func(_ *db.DBClient, request *pb.QueryPublishItemRequest) (*pb.QueryPublishItemData, error) {
		return &pb.QueryPublishItemData{
			List: []*pb.PublishItem{
				{
					ID: 1,
				},
			},
		}, nil
	})
	defer pm1.Unpatch()
	s := &PublishItemService{
		db: dbClient,
		p: &provider{
			Cfg: &config{
				SiteUrl: "https://erda.cloud",
			},
		},
	}
	_, err := s.QueryPublishItem(context.Background(), &pb.QueryPublishItemRequest{})
	assert.NoError(t, err)
}

func Test_checkPublishItemExsit(t *testing.T) {
	type args struct {
		publishItemID int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exist",
			args: args{
				publishItemID: 1,
			},
			wantErr: false,
		},
		{
			name: "not exist",
			args: args{
				publishItemID: 2,
			},
			wantErr: true,
		},
	}
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItem", func(_ *db.DBClient, id int64) (*db.PublishItem, error) {
		if id == 1 {
			return &db.PublishItem{}, nil
		}
		return nil, fmt.Errorf("not found")
	})
	defer pm1.Unpatch()
	s := &PublishItemService{
		db: dbClient,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := s.checkPublishItemExsit(tt.args.publishItemID); (err != nil) != tt.wantErr {
				t.Errorf("checkPublishItemExsit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updatePublishItemImpl(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItem", func(_ *db.DBClient, id int64) (*db.PublishItem, error) {
		return &db.PublishItem{}, nil
	})
	defer pm1.Unpatch()
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublishItem", func(_ *db.DBClient, publishItem *db.PublishItem) error {
		return nil
	})
	defer pm2.Unpatch()

	s := &PublishItemService{
		db: dbClient,
	}
	err := s.updatePublishItemImpl(&pb.UpdatePublishItemRequest{})
	assert.NoError(t, err)
}

func Test_getOrgIDFromContext(t *testing.T) {
	validCtx := context.WithValue(context.Background(), "org-id", 1)
	s := &PublishItemService{}
	_, err := s.getOrgIDFromContext(validCtx)
	assert.NoError(t, err)

	invalidCtx := context.WithValue(context.Background(), "org-id", "a")
	_, err = s.getOrgIDFromContext(invalidCtx)
	assert.NoError(t, err)
}

func Test_getPublishItemId(t *testing.T) {
	validVars := map[string]string{
		"publishItemId": "1",
	}
	publishID, err := getPublishItemId(validVars)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), publishID)

	invalidVars := map[string]string{
		"publishItemId": "a",
	}
	publishID, err = getPublishItemId(invalidVars)
	assert.Error(t, err)
	assert.Equal(t, int64(0), publishID)
}

func Test_getPermissionHeader(t *testing.T) {
	r := &http.Request{
		Header: map[string][]string{
			"Org-ID": {"1"},
		},
	}
	_, err := getPermissionHeader(r)
	assert.NoError(t, err)
}
