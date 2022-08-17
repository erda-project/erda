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
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/i18n"
)

func Test_publicReleaseVersion(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublicVersionByID", func(_ *db.DBClient, versionID int64, fileds map[string]interface{}) error {
		return nil
	})
	defer pm1.Unpatch()
	locale := &i18n.LocaleResource{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(locale), "Get", func(_ *i18n.LocaleResource, key string, defaults ...string) string {
		return ""
	})
	defer pm2.Unpatch()

	type arg struct {
		total    int
		req      *pb.UpdatePublishItemVersionStatesRequset
		versions []db.PublishItemVersion
	}
	tests := []struct {
		name    string
		wantErr bool
		arg     arg
	}{
		{
			name:    "total 0",
			wantErr: false,
			arg: arg{
				total: 0,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: nil,
			},
		},
		{
			name:    "total 1",
			wantErr: true,
			arg: arg{
				total: 1,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: []db.PublishItemVersion{
					{
						BaseModel: dbengine.BaseModel{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name:    "total 2",
			wantErr: false,
			arg: arg{
				total: 2,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: []db.PublishItemVersion{
					{
						BaseModel: dbengine.BaseModel{
							ID: 1,
						},
					},
					{
						BaseModel: dbengine.BaseModel{
							ID: 2,
						},
					},
				},
			},
		},
	}
	s := &PublishItemService{
		db: dbClient,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.publicReleaseVersion(tt.arg.total, tt.arg.versions, tt.arg.req, locale); (err != nil) != tt.wantErr {
				t.Errorf("publicReleaseVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_unPublicReleaseVersion(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublicVersionByID", func(_ *db.DBClient, versionID int64, fileds map[string]interface{}) error {
		return nil
	})
	defer pm1.Unpatch()

	s := &PublishItemService{
		db: dbClient,
	}
	locale := &i18n.LocaleResource{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(locale), "Get", func(_ *i18n.LocaleResource, key string, defaults ...string) string {
		return ""
	})
	defer pm2.Unpatch()
	err := s.unPublicReleaseVersion(1, []db.PublishItemVersion{
		{
			PublishItemID: 1,
		},
	}, &pb.UpdatePublishItemVersionStatesRequset{PublishItemID: 1}, locale)
	assert.NoError(t, err)
}

func Test_publicBetaVersion(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublicVersionByID", func(_ *db.DBClient, versionID int64, fileds map[string]interface{}) error {
		return nil
	})
	defer pm1.Unpatch()
	locale := &i18n.LocaleResource{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(locale), "Get", func(_ *i18n.LocaleResource, key string, defaults ...string) string {
		return ""
	})
	defer pm2.Unpatch()

	type arg struct {
		total    int
		req      *pb.UpdatePublishItemVersionStatesRequset
		versions []db.PublishItemVersion
	}
	tests := []struct {
		name    string
		wantErr bool
		arg     arg
	}{
		{
			name:    "total 0",
			wantErr: false,
			arg: arg{
				total: 0,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: nil,
			},
		},
		{
			name:    "total 1",
			wantErr: false,
			arg: arg{
				total: 1,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: []db.PublishItemVersion{
					{
						BaseModel: dbengine.BaseModel{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name:    "total 2",
			wantErr: false,
			arg: arg{
				total: 2,
				req: &pb.UpdatePublishItemVersionStatesRequset{
					PublishItemID:        1,
					PublishItemVersionID: 1,
				},
				versions: []db.PublishItemVersion{
					{
						BaseModel: dbengine.BaseModel{
							ID: 1,
						},
					},
					{
						BaseModel: dbengine.BaseModel{
							ID: 2,
						},
					},
				},
			},
		},
	}
	s := &PublishItemService{
		db: dbClient,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.publicBetaVersion(tt.arg.total, tt.arg.versions, tt.arg.req, locale); (err == nil) != tt.wantErr {
				t.Errorf("publicReleaseVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
