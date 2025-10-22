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

package akutil

import (
	"errors"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
)

type mockRequest struct {
	ClientId      string
	ClientTokenId string
}

func TestAutoCheckAndSetClientInfoSetsFields(t *testing.T) {
	req := &mockRequest{}

	clientID := "client-id"
	tokenID := "token-id"

	if err := AutoCheckAndSetClientInfo(clientID, tokenID, req, false); err != nil {
		t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
	}
	if req.ClientId != clientID {
		t.Fatalf("expected ClientId to be set to %q, got %q", clientID, req.ClientId)
	}
	if req.ClientTokenId != tokenID {
		t.Fatalf("expected ClientTokenId to be set to %q, got %q", tokenID, req.ClientTokenId)
	}
}

func TestAutoCheckAndSetClientInfoSkipSet(t *testing.T) {
	req := &mockRequest{}

	if err := AutoCheckAndSetClientInfo("client-id", "token-id", req, true); err != nil {
		t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
	}
	if req.ClientId != "" {
		t.Fatalf("expected ClientId to remain empty when skipSet is true, got %q", req.ClientId)
	}
	if req.ClientTokenId != "" {
		t.Fatalf("expected ClientTokenId to remain empty when skipSet is true, got %q", req.ClientTokenId)
	}
}

func TestAutoCheckAndSetClientInfoClientMismatch(t *testing.T) {
	req := &mockRequest{ClientId: "other"}

	err := AutoCheckAndSetClientInfo("client-id", "", req, false)
	if !errors.Is(err, handlers.ErrClientIdParamMismatch) {
		t.Fatalf("expected ErrClientIdParamMismatch, got %v", err)
	}
}

func TestAutoCheckAndSetClientInfoTokenMismatch(t *testing.T) {
	req := &mockRequest{ClientId: "client-id", ClientTokenId: "other-token"}

	err := AutoCheckAndSetClientInfo("client-id", "token-id", req, false)
	if !errors.Is(err, handlers.ErrTokenIdParamMismatch) {
		t.Fatalf("expected ErrTokenIdParamMismatch, got %v", err)
	}
}

func TestAutoCheckAndSetClientInfoFieldVariants(t *testing.T) {
	t.Run("ClientID field", func(t *testing.T) {
		req := &struct {
			ClientID string
		}{}

		if err := AutoCheckAndSetClientInfo("client-id", "", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.ClientID != "client-id" {
			t.Fatalf("expected ClientID to be set to %q, got %q", "client-id", req.ClientID)
		}
	})

	t.Run("ClientTokenID field", func(t *testing.T) {
		req := &struct {
			ClientTokenID string
		}{}

		if err := AutoCheckAndSetClientInfo("", "token-id", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.ClientTokenID != "token-id" {
			t.Fatalf("expected ClientTokenID to be set to %q, got %q", "token-id", req.ClientTokenID)
		}
	})

	t.Run("Client_Id field", func(t *testing.T) {
		req := &struct {
			Client_Id string
		}{}

		if err := AutoCheckAndSetClientInfo("client-id", "", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.Client_Id != "client-id" {
			t.Fatalf("expected Client_Id to be set to %q, got %q", "client-id", req.Client_Id)
		}
	})

	t.Run("Client_Token_Id field", func(t *testing.T) {
		req := &struct {
			Client_Token_Id string
		}{}

		if err := AutoCheckAndSetClientInfo("", "token-id", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.Client_Token_Id != "token-id" {
			t.Fatalf("expected Client_Token_Id to be set to %q, got %q", "token-id", req.Client_Token_Id)
		}
	})

	t.Run("CLIENTID field", func(t *testing.T) {
		req := &struct {
			CLIENTID string
		}{}

		if err := AutoCheckAndSetClientInfo("client-id", "", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.CLIENTID != "client-id" {
			t.Fatalf("expected CLIENTID to be set to %q, got %q", "client-id", req.CLIENTID)
		}
	})

	t.Run("CLIENTTOKENID field", func(t *testing.T) {
		req := &struct {
			CLIENTTOKENID string
		}{}

		if err := AutoCheckAndSetClientInfo("", "token-id", req, false); err != nil {
			t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
		}
		if req.CLIENTTOKENID != "token-id" {
			t.Fatalf("expected CLIENTTOKENID to be set to %q, got %q", "token-id", req.CLIENTTOKENID)
		}
	})
}

func TestAutoCheckAndSetClientInfoNilRequest(t *testing.T) {
	if err := AutoCheckAndSetClientInfo("client-id", "token-id", nil, false); err != nil {
		t.Fatalf("expected nil error for nil request, got %v", err)
	}
}

func TestAutoCheckAndSetClientInfoUnexportedFields(t *testing.T) {
	req := &struct {
		clientId      string
		clientTokenId string
	}{}

	if err := AutoCheckAndSetClientInfo("client-id", "token-id", req, false); err != nil {
		t.Fatalf("AutoCheckAndSetClientInfo returned error: %v", err)
	}
	if req.clientId != "" {
		t.Fatalf("expected unexported clientId to remain empty, got %q", req.clientId)
	}
	if req.clientTokenId != "" {
		t.Fatalf("expected unexported clientTokenId to remain empty, got %q", req.clientTokenId)
	}
}
