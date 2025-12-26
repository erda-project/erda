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

package admin_token

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitAdminAuthToken(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "admin_token_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tokenPath := filepath.Join(tmpDir, "auth_token")

	// 1. First init
	token1, err := InitAdminAuthToken(tokenPath)
	if err != nil {
		t.Errorf("InitAdminAuthToken failed: %v", err)
	}
	if token1 == "" {
		t.Error("expected non-empty token")
	}

	// 2. Check file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("token file should exist")
	}

	// 3. Second init (should return same token, but actually once.Do makes it same token in same process)
	token2, err := InitAdminAuthToken(tokenPath)
	if err != nil {
		t.Errorf("InitAdminAuthToken failed: %v", err)
	}
	if token1 != token2 {
		t.Errorf("expected same token, got %s and %s", token1, token2)
	}

	// 4. Validate
	if !ValidateAdminAuthToken(token1) {
		t.Error("ValidateAdminAuthToken should return true for correct token")
	}
	if ValidateAdminAuthToken("wrong") {
		t.Error("ValidateAdminAuthToken should return false for wrong token")
	}
	if GetAdminAuthToken() != token1 {
		t.Errorf("expected GetAdminAuthToken() to return %s", token1)
	}
}
