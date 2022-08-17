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
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/pkg/i18n"
)

func TestPublishItemVersion(t *testing.T) {
	type arg struct {
		req *pb.CreatePublishItemVersionRequest
	}
	listValue, err := structpb.NewList([]interface{}{"mobile"})
	assert.NoError(t, err)
	tests := []struct {
		name string
		args arg
	}{
		{
			name: "with release id",
			args: arg{
				req: &pb.CreatePublishItemVersionRequest{
					ReleaseID:     "1",
					PublishItemID: 1,
				},
			},
		},
		{
			name: "with app id",
			args: arg{
				req: &pb.CreatePublishItemVersionRequest{
					AppID:         1,
					PublishItemID: 2,
					MobileType:    string(apistructs.ResourceTypeH5),
					H5VersionInfo: &pb.H5VersionInfo{
						TargetMobiles: map[string]*structpb.Value{
							"h5": structpb.NewListValue(listValue),
						},
					},
				},
			},
		},
	}
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItem", func(_ *db.DBClient, id int64) (*db.PublishItem, error) {
		return &db.PublishItem{
			PublisherID: 1,
		}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItemVersionByName", func(_ *db.DBClient, orgId int64, itemID int64,
		mobileType apistructs.ResourceType, versionInfo *pb.VersionInfo) (*db.PublishItemVersion, error) {
		if itemID == 1 {
			return nil, fmt.Errorf("item id is 1")
		}
		return &db.PublishItemVersion{}, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "CreatePublishItemVersion", func(_ *db.DBClient, item *db.PublishItemVersion) error {
		return nil
	})
	defer pm3.Unpatch()

	pm4 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublishItemVersion", func(_ *db.DBClient, item *db.PublishItemVersion) error {
		return nil
	})
	defer pm4.Unpatch()

	bdl := bundle.New()
	pm5 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetRelease", func(_ *bundle.Bundle, releaseID string) (*apistructs.ReleaseGetResponseData, error) {
		return &apistructs.ReleaseGetResponseData{
			ReleaseID: releaseID,
		}, nil
	})
	defer pm5.Unpatch()

	pm6 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp", func(_ *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			ID: id,
		}, nil
	})
	defer pm6.Unpatch()

	i := &PublishItemService{
		bdl: bdl,
		db:  dbClient,
	}
	for _, tt := range tests {
		_, err := i.PublishItemVersion(tt.args.req)
		assert.NoError(t, err)
	}
}

func TestGetPublishItemDistribution(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItem", func(_ *db.DBClient, id int64) (*db.PublishItem, error) {
		return &db.PublishItem{
			PublisherID: 1,
			Type:        apistructs.PublishItemTypeMobile,
		}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "QueryPublishItemVersions", func(_ *db.DBClient, request *pb.QueryPublishItemVersionRequest) (*pb.QueryPublishItemVersionData, error) {
		return &pb.QueryPublishItemVersionData{
			Total: 1,
			List: []*pb.PublishItemVersion{
				{
					ID: 1,
				},
			},
		}, nil
	})
	defer pm2.Unpatch()

	i := &PublishItemService{
		db: dbClient,
	}

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(i), "GrayDistribution", func(_ *PublishItemService, w http.ResponseWriter, r *http.Request, publisherItem db.PublishItem, distribution *pb.PublishItemDistributionData, mobileType apistructs.ResourceType, packageName string) error {
		return nil
	})
	defer pm3.Unpatch()

	_, err := i.GetPublishItemDistribution(1, "android", "android", &mock.MockHTTPResponseWriter{}, &http.Request{})
	assert.NoError(t, err)
}

func TestGetPublicPublishItemVersionImpl(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublicVersion", func(_ *db.DBClient, itemID int64, mobileType apistructs.ResourceType, packageName string) (int, []db.PublishItemVersion, error) {
		return 2, []db.PublishItemVersion{
			{
				VersionStates: string(apistructs.PublishItemReleaseVersion),
			},
			{
				VersionStates: string(apistructs.PublishItemBetaVersion),
			},
		}, nil
	})
	defer pm1.Unpatch()

	i := &PublishItemService{
		db: dbClient,
	}
	_, err := i.GetPublicPublishItemVersionImpl(1, "android", "")
	assert.NoError(t, err)
}

func TestPublicPublishItemVersion(t *testing.T) {
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItemVersion", func(_ *db.DBClient, id int64) (*db.PublishItemVersion, error) {
		return &db.PublishItemVersion{}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublicVersion", func(_ *db.DBClient, itemID int64, mobileType apistructs.ResourceType, packageName string) (int, []db.PublishItemVersion, error) {
		return 1, []db.PublishItemVersion{
			{
				VersionStates: string(apistructs.PublishItemReleaseVersion),
			},
		}, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "UpdatePublicVersionByID", func(_ *db.DBClient, versionID int64, fileds map[string]interface{}) error {
		return nil
	})
	defer pm3.Unpatch()

	locale := &i18n.LocaleResource{}
	pm4 := monkey.PatchInstanceMethod(reflect.TypeOf(locale), "Get", func(_ *i18n.LocaleResource, key string, defaults ...string) string {
		return ""
	})
	defer pm4.Unpatch()

	i := &PublishItemService{
		db: dbClient,
	}
	err := i.PublicPublishItemVersion(&pb.UpdatePublishItemVersionStatesRequset{
		VersionStates: string(apistructs.PublishItemReleaseVersion),
	}, locale)
	assert.NoError(t, err)
}

func TestGetPublicPublishItemLaststVersion(t *testing.T) {
	bdl := bundle.New()
	dbClient := &db.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "QueryAppPublishItemRelations", func(_ *bundle.Bundle, req *apistructs.QueryAppPublishItemRelationRequest) (*apistructs.QueryAppPublishItemRelationResponse, error) {
		return &apistructs.QueryAppPublishItemRelationResponse{
			Data: []apistructs.AppPublishItemRelation{
				{
					PublishItemID: 1,
				},
			},
		}, nil
	})
	defer pm1.Unpatch()
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItem", func(_ *db.DBClient, id int64) (*db.PublishItem, error) {
		return &db.PublishItem{
			PublisherID: 1,
		}, nil
	})
	defer pm2.Unpatch()
	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPublishItemVersionByName", func(_ *db.DBClient, orgId int64, itemID int64,
		mobileType apistructs.ResourceType, versionInfo *pb.VersionInfo) (*db.PublishItemVersion, error) {
		return &db.PublishItemVersion{
			PublishItemID: 1,
		}, nil
	})
	defer pm3.Unpatch()
	s := &PublishItemService{
		bdl: bdl,
		db:  dbClient,
	}
	pm4 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetPublishItemDistribution", func(_ *PublishItemService, id int64, mobileType apistructs.ResourceType, packageName string, w http.ResponseWriter, r *http.Request) (*pb.PublishItemDistributionData, error) {
		return &pb.PublishItemDistributionData{
			Name: "publish-item",
		}, nil
	})
	defer pm4.Unpatch()
	pm5 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetPublicPublishItemVersionImpl", func(_ *PublishItemService, itemID int64, mobileType string, packageName string) (*pb.QueryPublishItemVersionData, error) {
		return &pb.QueryPublishItemVersionData{
			List: []*pb.PublishItemVersion{
				{
					ID: 1,
				},
				{
					ID: 2,
				},
				{
					ID: 3,
				},
			},
		}, nil
	})
	defer pm5.Unpatch()
	_, err := s.GetPublicPublishItemLaststVersion(mock.NewMockHTTPResponseWriter(), &http.Request{}, pb.GetPublishItemLatestVersionRequest{
		ForceBetaH5: true,
		CurrentH5Info: []*pb.VersionInfo{
			{
				Version: "1.0",
			},
		},
	})
	assert.NoError(t, err)
}

func Test_getNewestVersion(t *testing.T) {
	version := getNewestVersion([]*db.PublishItemVersion{{
		Version: "1.0",
	}, {
		Version: "2.0",
	},
	}...)
	assert.Equal(t, "2.0", version.Version)
}

func Test_getLogo(t *testing.T) {
	logo := getLogo(`[
    {
        "type": "android",
        "meta": {
            "logo": "https://erda.cloud"
        }
    }
]`)
	assert.Equal(t, "https://erda.cloud", logo)
}
