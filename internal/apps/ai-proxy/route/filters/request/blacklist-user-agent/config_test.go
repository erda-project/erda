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

package blacklist_user_agent

import (
	"reflect"
	"testing"
)

func TestConfigFileTags(t *testing.T) {
	clientTokenField, ok := reflect.TypeOf(ClientTokenConfig{}).FieldByName("Blacklist")
	if !ok {
		t.Fatal("Blacklist field not found in ClientTokenConfig")
	}
	if got := clientTokenField.Tag.Get("file"); got != "blacklist" {
		t.Fatalf("unexpected client_token file tag: %q", got)
	}

	clientField, ok := reflect.TypeOf(ClientConfig{}).FieldByName("Blacklist")
	if !ok {
		t.Fatal("Blacklist field not found in ClientConfig")
	}
	if got := clientField.Tag.Get("file"); got != "blacklist" {
		t.Fatalf("unexpected client file tag: %q", got)
	}
}

func TestSetConfig_NormalizesBlacklist(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })

	SetConfig(Config{
		ClientToken: ClientTokenConfig{
			Blacklist: []string{" openclaw ", "", " CoDeX "},
		},
		Client: ClientConfig{
			Blacklist: nil,
		},
	})

	cfg := getConfig()
	if len(cfg.ClientToken.Blacklist) != 2 || cfg.ClientToken.Blacklist[0] != "openclaw" || cfg.ClientToken.Blacklist[1] != "codex" {
		t.Fatalf("unexpected client_token blacklist: %#v", cfg.ClientToken.Blacklist)
	}
	if len(cfg.Client.Blacklist) != 0 {
		t.Fatalf("expected empty client blacklist, got %#v", cfg.Client.Blacklist)
	}
}
