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
	if !errors.Is(err, handlers.ErrAkNotMatch) {
		t.Fatalf("expected ErrAkNotMatch, got %v", err)
	}
}

func TestAutoCheckAndSetClientInfoTokenMismatch(t *testing.T) {
	req := &mockRequest{ClientId: "client-id", ClientTokenId: "other-token"}

	err := AutoCheckAndSetClientInfo("client-id", "token-id", req, false)
	if !errors.Is(err, handlers.ErrTokenNotMatch) {
		t.Fatalf("expected ErrTokenNotMatch, got %v", err)
	}
}
