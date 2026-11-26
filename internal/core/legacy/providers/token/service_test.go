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

package token

import (
	"context"
	"testing"

	"github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	tokenstore "github.com/erda-project/erda/pkg/oauth2/tokenstore/mysqltokenstore"
)

func TestToModelToken(t *testing.T) {
	type args struct {
		userID string
		req    *pb.CreateTokenRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				req: &pb.CreateTokenRequest{
					Type: "AccessKey",
				},
				userID: "2",
			},
			wantErr: false,
		},
		{
			args: args{
				req: &pb.CreateTokenRequest{
					Type:  "PAT",
					Scope: "org",
				},
			},
			wantErr: false,
		},
		{
			args: args{
				req: &pb.CreateTokenRequest{
					Type: "PAT",
				},
			},
			wantErr: true,
		},
		{
			args: args{
				req: &pb.CreateTokenRequest{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToModelToken(tt.args.userID, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToModelToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type fakeTokenDao struct {
	queryReq *pb.QueryTokensRequest
	items    []tokenstore.TokenStoreItem
}

func (d *fakeTokenDao) QueryToken(_ context.Context, req *pb.QueryTokensRequest) ([]tokenstore.TokenStoreItem, int64, error) {
	copied := *req
	d.queryReq = &copied
	return d.items, int64(len(d.items)), nil
}

func (d *fakeTokenDao) CreateToken(_ context.Context, obj tokenstore.TokenStoreItem) (*tokenstore.TokenStoreItem, error) {
	return &obj, nil
}

func (d *fakeTokenDao) GetToken(_ context.Context, _ *pb.GetTokenRequest) (*tokenstore.TokenStoreItem, error) {
	return &d.items[0], nil
}

func (d *fakeTokenDao) UpdateToken(_ context.Context, _ *pb.UpdateTokenRequest) error {
	return nil
}

func (d *fakeTokenDao) DeleteToken(_ context.Context, _ *pb.DeleteTokenRequest) error {
	return nil
}

type fakeSysAdminChecker struct {
	admin bool
}

func (c fakeSysAdminChecker) IsSysAdmin(_ string) (bool, error) {
	return c.admin, nil
}

func TestQueryTokensRequiresExplicitScopeAndMasksExternalResponses(t *testing.T) {
	item := tokenstore.TokenStoreItem{
		ID:        "token-1",
		AccessKey: "abcdefghijklmnopqrstuvwxyz123456",
		SecretKey: "secret-value",
		CreatorID: "user-a",
	}
	dao := &fakeTokenDao{items: []tokenstore.TokenStoreItem{item}}
	service := &TokenService{dao: dao, sysAdminChecker: fakeSysAdminChecker{}}

	externalCtx := apis.WithUserIDContext(context.Background(), "user-a")
	resp, err := service.QueryTokens(externalCtx, &pb.QueryTokensRequest{Scope: "org", ScopeId: "1"})
	if err != nil {
		t.Fatalf("QueryTokens() error = %v", err)
	}
	if dao.queryReq.CreatorId != "user-a" {
		t.Errorf("creator filter = %q, want user-a", dao.queryReq.CreatorId)
	}
	if got := resp.Data[0].AccessKey; got == item.AccessKey {
		t.Errorf("external response returned unmasked access key")
	}
	if got := resp.Data[0].SecretKey; got != "" {
		t.Errorf("external response returned secret key %q", got)
	}

	internalCtx := apis.WithInternalClientContext(context.Background(), "msp")
	resp, err = service.QueryTokens(internalCtx, &pb.QueryTokensRequest{Scope: "msp_env", ScopeId: "1"})
	if err != nil {
		t.Fatalf("internal QueryTokens() error = %v", err)
	}
	if got := resp.Data[0].AccessKey; got != item.AccessKey {
		t.Errorf("internal response access key = %q, want full value", got)
	}
}

func TestQueryTokensRejectsUnboundedAndCrossUserQueries(t *testing.T) {
	service := &TokenService{dao: &fakeTokenDao{}, sysAdminChecker: fakeSysAdminChecker{}}

	if _, err := service.QueryTokens(apis.WithInternalClientContext(context.Background(), "msp"), &pb.QueryTokensRequest{}); err == nil {
		t.Errorf("internal unbounded query should be rejected")
	}
	if _, err := service.QueryTokens(context.Background(), &pb.QueryTokensRequest{}); err == nil {
		t.Errorf("identity-less unbounded query should be rejected")
	}
	if _, err := service.QueryTokens(apis.WithUserIDContext(context.Background(), "user-a"), &pb.QueryTokensRequest{CreatorId: "user-b"}); err == nil {
		t.Errorf("cross-user query should be rejected")
	}
}

func TestGetTokenMasksExternalAndAllowsInternal(t *testing.T) {
	item := tokenstore.TokenStoreItem{
		ID:        "token-1",
		AccessKey: "abcdefghijklmnopqrstuvwxyz123456",
		SecretKey: "secret-value",
		CreatorID: "user-a",
	}
	service := &TokenService{dao: &fakeTokenDao{items: []tokenstore.TokenStoreItem{item}}, sysAdminChecker: fakeSysAdminChecker{}}

	external, err := service.GetToken(apis.WithUserIDContext(context.Background(), "user-a"), &pb.GetTokenRequest{Id: item.ID})
	if err != nil {
		t.Fatalf("external GetToken() error = %v", err)
	}
	if external.Data.AccessKey == item.AccessKey || external.Data.SecretKey != "" {
		t.Errorf("external GetToken() returned unmasked credentials")
	}

	internal, err := service.GetToken(apis.WithInternalClientContext(context.Background(), "msp"), &pb.GetTokenRequest{Id: item.ID})
	if err != nil {
		t.Fatalf("internal GetToken() error = %v", err)
	}
	if internal.Data.AccessKey != item.AccessKey || internal.Data.SecretKey != item.SecretKey {
		t.Errorf("internal GetToken() did not return full credentials")
	}
}
