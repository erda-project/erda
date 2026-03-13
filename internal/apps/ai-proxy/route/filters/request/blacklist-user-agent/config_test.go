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

func TestConfigEnvTags(t *testing.T) {
	clientTokenField, ok := reflect.TypeOf(ClientTokenConfig{}).FieldByName("Blacklist")
	if !ok {
		t.Fatal("Blacklist field not found in ClientTokenConfig")
	}
	if got := clientTokenField.Tag.Get("env"); got != "AI_PROXY_BLACKLIST_USER_AGENT_FOR_CLIENT_TOKEN" {
		t.Fatalf("unexpected client_token env tag: %q", got)
	}

	clientField, ok := reflect.TypeOf(ClientConfig{}).FieldByName("Blacklist")
	if !ok {
		t.Fatal("Blacklist field not found in ClientConfig")
	}
	if got := clientField.Tag.Get("env"); got != "AI_PROXY_BLACKLIST_USER_AGENT_FOR_CLIENT" {
		t.Fatalf("unexpected client env tag: %q", got)
	}
}
