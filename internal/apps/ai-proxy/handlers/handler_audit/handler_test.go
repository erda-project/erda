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

package handler_audit

import (
	"context"
	"testing"

	auditpb "github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func newCtx() context.Context {
	return ctxhelper.InitCtxMapIfNeed(context.Background())
}

func TestCheckAndFillAuth_Admin_AllowsMissingXRequestIdOrCallId(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, true)
	req := &auditpb.AuditPagingRequest{}
	isAdmin, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isAdmin {
		t.Fatalf("expected isAdmin=true")
	}
	if req.AuthKey != "" {
		t.Fatalf("admin path should not override authKey, got %q", req.AuthKey)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err != nil {
		t.Fatalf("admin should bypass identifier requirement: %v", err)
	}
}

func TestCheckAndFillAuth_ClientAK_MissingXRequestIdOrCallId_Error(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, false)
	ctxhelper.PutClient(ctx, &clientpb.Client{AccessKeyId: "AK1"})
	req := &auditpb.AuditPagingRequest{}
	_, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error from checkAndFillAuth: %v", err)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err == nil {
		t.Fatalf("expected error for missing x_request_id or call_id, got nil")
	}
}

func TestCheckAndFillAuth_ClientAK_WithXRequestId_SetsAuthKey(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, false)
	ctxhelper.PutClient(ctx, &clientpb.Client{AccessKeyId: "AK1"})
	req := &auditpb.AuditPagingRequest{XRequestId: "RID1"}
	isAdmin, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isAdmin {
		t.Fatalf("expected isAdmin=false")
	}
	if req.AuthKey != "AK1" {
		t.Fatalf("expected authKey to be set to AK, got %q", req.AuthKey)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err != nil {
		t.Fatalf("unexpected error from requireXRequestIdOrCallIdForNonAdmin: %v", err)
	}
}

func TestCheckAndFillAuth_ClientToken_PrefersTokenAndRequiresXRequestIdOrCallId(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, false)
	ctxhelper.PutClientToken(ctx, &clienttokenpb.ClientToken{Token: "t_token1"})
	ctxhelper.PutClient(ctx, &clientpb.Client{AccessKeyId: "AK1"})

	// missing x_request_id and call_id -> error only from requireXRequestIdOrCallIdForNonAdmin
	req := &auditpb.AuditPagingRequest{}
	_, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error from checkAndFillAuth: %v", err)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err == nil {
		t.Fatalf("expected error for missing x_request_id or call_id when using token")
	}

	// with x_request_id -> use token
	req = &auditpb.AuditPagingRequest{XRequestId: "RID2"}
	_, err = checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.AuthKey != "t_token1" {
		t.Fatalf("expected authKey to be token, got %q", req.AuthKey)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err != nil {
		t.Fatalf("unexpected error from requireXRequestIdOrCallIdForNonAdmin: %v", err)
	}
}

func TestCheckAndFillAuth_ClientAK_WithCallId_SetsAuthKey(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, false)
	ctxhelper.PutClient(ctx, &clientpb.Client{AccessKeyId: "AK1"})
	req := &auditpb.AuditPagingRequest{CallId: "CID1"}
	isAdmin, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isAdmin {
		t.Fatalf("expected isAdmin=false")
	}
	if req.AuthKey != "AK1" {
		t.Fatalf("expected authKey to be set to AK, got %q", req.AuthKey)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err != nil {
		t.Fatalf("unexpected error from requireXRequestIdOrCallIdForNonAdmin: %v", err)
	}
}

func TestCheckAndFillAuth_ClientToken_WithCallId_PrefersToken(t *testing.T) {
	ctx := newCtx()
	ctxhelper.PutIsAdmin(ctx, false)
	ctxhelper.PutClientToken(ctx, &clienttokenpb.ClientToken{Token: "t_token1"})
	ctxhelper.PutClient(ctx, &clientpb.Client{AccessKeyId: "AK1"})

	// with call_id -> use token
	req := &auditpb.AuditPagingRequest{CallId: "CID2"}
	_, err := checkAndFillAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.AuthKey != "t_token1" {
		t.Fatalf("expected authKey to be token, got %q", req.AuthKey)
	}
	if err := requireXRequestIdOrCallIdForNonAdmin(ctx, req); err != nil {
		t.Fatalf("unexpected error from requireXRequestIdOrCallIdForNonAdmin: %v", err)
	}
}
