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
	clientTokenField, ok := reflect.TypeOf(ClientTokenConfig{}).FieldByName("BlacklistStr")
	if !ok {
		t.Fatal("BlacklistStr field not found in ClientTokenConfig")
	}
	if got := clientTokenField.Tag.Get("file"); got != "blacklist_str" {
		t.Fatalf("unexpected client_token file tag: %q", got)
	}

	clientField, ok := reflect.TypeOf(ClientConfig{}).FieldByName("BlacklistStr")
	if !ok {
		t.Fatal("BlacklistStr field not found in ClientConfig")
	}
	if got := clientField.Tag.Get("file"); got != "blacklist_str" {
		t.Fatalf("unexpected client file tag: %q", got)
	}
}

func TestSetConfig_NormalizesBlacklistStr(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })

	SetConfig(Config{
		ClientToken: ClientTokenConfig{
			BlacklistStr: " openclaw, CoDeX, cursor ,, ",
		},
		Client: ClientConfig{
			BlacklistStr: " codex, openclaw, cursor ",
		},
	})

	cfg := getConfig()
	if len(cfg.ClientToken.Blacklist) != 3 || cfg.ClientToken.Blacklist[0] != "openclaw" || cfg.ClientToken.Blacklist[1] != "codex" || cfg.ClientToken.Blacklist[2] != "cursor" {
		t.Fatalf("unexpected client_token blacklist: %#v", cfg.ClientToken.Blacklist)
	}
	if len(cfg.Client.Blacklist) != 3 || cfg.Client.Blacklist[0] != "codex" || cfg.Client.Blacklist[1] != "openclaw" || cfg.Client.Blacklist[2] != "cursor" {
		t.Fatalf("unexpected client blacklist: %#v", cfg.Client.Blacklist)
	}
}

func TestSetGeneralRules_NormalizesHeadersAndPrompts(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })

	SetGeneralRules(" Claude Code ;;; X-Code;Agent ;;; ;; ", " You are OpenCode ;;; You are Claude; Code ;;; ;; ")

	cfg := getGeneralRules()
	if len(cfg.Headers) != 2 || cfg.Headers[0] != "claude code" || cfg.Headers[1] != "x-code;agent" {
		t.Fatalf("unexpected general header rules: %#v", cfg.Headers)
	}
	if len(cfg.Prompts) != 2 || cfg.Prompts[0] != "you are opencode" || cfg.Prompts[1] != "you are claude; code" {
		t.Fatalf("unexpected general prompt rules: %#v", cfg.Prompts)
	}
}
