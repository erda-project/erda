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

package org

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/org/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

func TestGetOrgByDomainAndOrgName(t *testing.T) {
	org := &db.Org{Name: "org0"}
	dbClient := &db.DBClient{}
	orgByName := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetOrgByName", func(_ *db.DBClient, orgName string) (*db.Org, error) {
		if orgName == "org0" {
			return org, nil
		} else {
			return nil, db.ErrNotFoundOrg
		}
	})
	defer orgByName.Unpatch()

	p := &provider{dbClient: dbClient}
	orgByDomain := monkey.PatchInstanceMethod(reflect.TypeOf(p), "GetOrgByDomainFromDB", func(_ *provider, domain string) (*db.Org, error) {
		if domain == "org0" {
			return org, nil
		} else {
			return nil, nil
		}
	})
	defer orgByDomain.Unpatch()

	res, err := p.GetOrgByDomainAndOrgName("org0", "")
	require.NoError(t, err)
	assert.Equal(t, org, res)
	res, err = p.GetOrgByDomainAndOrgName("org0", "org1")
	require.NoError(t, err)
	assert.Equal(t, org, res)
	res, err = p.GetOrgByDomainAndOrgName("org2", "org1")
	require.NoError(t, err)
	assert.Equal(t, (*db.Org)(nil), res)
}

func TestOrgNameRetriever(t *testing.T) {
	var domains = []string{"erda-org.erda.cloud", "buzz-org.app.terminus.io", "fuzz.com"}
	var domainRoots = []string{"erda.cloud", "app.terminus.io"}
	assert.Equal(t, "erda", orgNameRetriever(domains[0], domainRoots[0]))
	assert.Equal(t, "buzz", orgNameRetriever(domains[1], domainRoots[1]))
	assert.Equal(t, "", orgNameRetriever(domains[2], domainRoots[0]))
}

func TestOrg_ListOrgs(t *testing.T) {
	type args struct {
		orgIDs []int64
		req    *pb.ListOrgRequest
		all    bool
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   []db.Org
		wantErr bool
	}{
		{
			name: "test_return_error",
			args: args{
				orgIDs: []int64{1},
				req:    &pb.ListOrgRequest{},
				all:    false,
			},
			want:    0,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "test_all",
			args: args{
				orgIDs: []int64{1},
				req:    &pb.ListOrgRequest{},
				all:    true,
			},
			want: 1,
			want1: []db.Org{
				{
					BaseModel: db.BaseModel{
						ID: 1,
					},
					Name: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "test_not_all",
			args: args{
				orgIDs: []int64{1},
				req:    &pb.ListOrgRequest{},
				all:    false,
			},
			want: 2,
			want1: []db.Org{
				{
					BaseModel: db.BaseModel{
						ID: 1,
					},
					Name: "2",
				},
			},
			wantErr: false,
		},
	}
	ctx := apis.WithOrgIDContext(apis.WithUserIDContext(context.Background(), "1"), "1")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				Cfg: &config{},
			}
			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "SearchByName", func(p *provider, name string, pageNo, pageSize int) (int, []db.Org, error) {
				if tt.wantErr {
					return 0, nil, fmt.Errorf("error")
				}
				return tt.want, tt.want1, nil
			})
			defer patch1.Unpatch()

			patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ListByIDsAndName", func(p *provider, orgIDs []int64, name string, pageNo, pageSize int) (int, []db.Org, error) {
				if tt.wantErr {
					return 0, nil, fmt.Errorf("error")
				}
				return tt.want, tt.want1, nil
			})
			defer patch2.Unpatch()

			got, got1, err := p.ListOrgs(ctx, tt.args.orgIDs, tt.args.req, tt.args.all)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOrgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ListOrgs() got = %v, want %v", got, tt.want)
			}
			dto, err := p.coverOrgsToDto(ctx, tt.want1)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !reflect.DeepEqual(got1, dto) {
				t.Errorf("ListOrgs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetAuditMessage(t *testing.T) {
	tt := []struct {
		org  db.Org
		req  pb.UpdateOrgRequest
		want *pb.AuditMessage
	}{
		{
			db.Org{DisplayName: "dice"},
			pb.UpdateOrgRequest{DisplayName: "erda"},
			&pb.AuditMessage{
				MessageZH: "组织名称由 dice 改为 erda ",
				MessageEN: "org name updated from dice to erda ",
			},
		},
		{
			db.Org{Locale: "en-US"},
			pb.UpdateOrgRequest{Locale: "zh-CN"},
			&pb.AuditMessage{
				MessageZH: "通知语言改为中文 ",
				MessageEN: "language updated to zh-CN ",
			},
		},
		{
			db.Org{IsPublic: true},
			pb.UpdateOrgRequest{IsPublic: false},
			&pb.AuditMessage{
				MessageZH: "改为私有组织 ",
				MessageEN: "org updated to private ",
			},
		},
		{
			db.Org{Logo: ""},
			pb.UpdateOrgRequest{Logo: "foo.png"},
			&pb.AuditMessage{
				MessageZH: "组织Logo发生变更 ",
				MessageEN: "org Logo changed ",
			},
		},
		{
			db.Org{Desc: ""},
			pb.UpdateOrgRequest{Desc: "bar"},
			&pb.AuditMessage{
				MessageZH: "组织描述信息发生变更 ",
				MessageEN: "org desc changed ",
			},
		},
		{
			db.Org{BlockoutConfig: db.BlockoutConfig{
				BlockDEV:   false,
				BlockTEST:  false,
				BlockStage: true,
				BlockProd:  true,
			}},
			pb.UpdateOrgRequest{BlockoutConfig: &pb.BlockoutConfig{
				BlockDev:   true,
				BlockTest:  true,
				BlockStage: false,
				BlockProd:  false,
			}},
			&pb.AuditMessage{
				MessageZH: "开发环境开启封网 测试环境开启封网 预发环境关闭封网 生产环境关闭封网 ",
				MessageEN: "block network opened in dev environment block network opened in test environment block network closed in staging environment block network closed in prod environment ",
			},
		},
		{
			db.Org{BlockoutConfig: db.BlockoutConfig{
				BlockDEV:   true,
				BlockTEST:  true,
				BlockStage: true,
				BlockProd:  true,
			},
				DisplayName: "dice", Locale: "en-US", IsPublic: true, Logo: "", Desc: ""},
			pb.UpdateOrgRequest{BlockoutConfig: &pb.BlockoutConfig{
				BlockDev:   true,
				BlockTest:  true,
				BlockStage: true,
				BlockProd:  true,
			},
				DisplayName: "dice", Locale: "en-US", IsPublic: true, Logo: "", Desc: ""},
			&pb.AuditMessage{
				MessageZH: "无信息变更",
				MessageEN: "no message changed",
			},
		},
	}
	for _, v := range tt {
		message := getAuditMessage(v.org, &v.req)
		if v.want.MessageEN != message.MessageEN {
			t.Error("fail")
		}
		if v.want.MessageZH != message.MessageZH {
			t.Error("fail")
		}

	}
}
