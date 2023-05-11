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

package audit_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func TestAudit_SetSessionId(t *testing.T) {
	var header = http.Header{
		"X-Erda-AI-Proxy-SessionId": []string{"mocked-session-id"},
	}
	f, err := audit.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.(*audit.Audit).SetSessionId(context.Background(), header); err != nil {
		t.Fatal(err)
	}
	if f.(*audit.Audit).Audit.SessionId != header.Get("X-Erda-AI-Proxy-SessionId") {
		t.Error("X-Erda-AI-Proxy-SessionId")
	}
}

func TestAudit_SetSource(t *testing.T) {
	var header = http.Header{
		"X-Erda-AI-Proxy-Source": []string{"mocked-session-id"},
	}
	f, _ := audit.New(nil)
	a := f.(*audit.Audit)
	if err := a.SetSource(context.Background(), header); err != nil {
		t.Fatal(err)
	}
	if a.Audit.Source != header.Get("X-Erda-AI-Proxy-Source") {
		t.Error("X-Erda-AI-Proxy-Source")
	}
}

func TestAudit_SetUserInfo(t *testing.T) {
	var m = map[string]string{
		"X-Erda-AI-Proxy-Name":            "mocked-name",
		"X-Erda-AI-Proxy-Phone":           "mocked-phone",
		"X-Erda-AI-Proxy-JobNumber":       "mocked-job-number",
		"X-Erda-AI-Proxy-Email":           "mocked-email",
		"X-Erda-AI-Proxy-DingTalkStaffID": "mocked-ding-id",
	}
	var header = make(http.Header)
	for k, v := range m {
		header.Set(k, base64.StdEncoding.EncodeToString([]byte(v)))
	}

	f, _ := audit.New(nil)
	a := f.(*audit.Audit)
	if err := a.SetUserInfo(context.Background(), header); err != nil {
		t.Fatal(err)
	}
	if a.Audit.Username != m["X-Erda-AI-Proxy-Name"] {
		t.Error("X-Erda-AI-Proxy-Name")
	}
	if a.Audit.PhoneNumber != m["X-Erda-AI-Proxy-Phone"] {
		t.Error("X-Erda-AI-Proxy-Phone")
	}
	if a.Audit.JobNumber != m["X-Erda-AI-Proxy-JobNumber"] {
		t.Error("X-Erda-AI-Proxy-JobNumber")
	}
	if a.Audit.Email != m["X-Erda-AI-Proxy-Email"] {
		t.Error("X-Erda-AI-Proxy-Email")
	}
	if a.Audit.DingtalkStaffId != m["X-Erda-AI-Proxy-DingTalkStaffID"] {
		t.Error("X-Erda-AI-Proxy-DingTalkStaffID")
	}
}

func TestAudit_SetChats(t *testing.T) {
	var a = &audit.Audit{
		DefaultResponseFilter: nil,
		Audit:                 new(models.AIProxyFilterAudit),
	}
	var cases = map[string]string{
		"X-Erda-AI-Proxy-ChatType":  "group",
		"X-Erda-AI-Proxy-ChatTitle": "erda-family",
		"X-Erda-AI-Proxy-ChatId":    "mocked-id",
	}
	var header = make(http.Header)
	for k, v := range cases {
		header.Set(k, base64.StdEncoding.EncodeToString([]byte(v)))
	}
	if err := a.SetChats(context.Background(), header); err != nil {
		t.Fatal(err)
	}
	if a.Audit.ChatType != cases["X-Erda-AI-Proxy-ChatType"] {
		t.Error("X-Erda-AI-Proxy-ChatType")
	}
	if a.Audit.ChatTitle != cases["X-Erda-AI-Proxy-ChatTitle"] {
		t.Error("X-Erda-AI-Proxy-ChatTitle")
	}
	if a.Audit.ChatId != cases["X-Erda-AI-Proxy-ChatId"] {
		t.Error("X-Erda-AI-Proxy-ChatTitle")
	}
}

func TestAudit_SetPrompt(t *testing.T) {
	f, err := audit.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	aud, ok := f.(*audit.Audit)
	if !ok {
		t.Fatal("it should be ok")
	}
	t.Run("POST /v1/chat/completions", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "http://ai.localhost/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"Your Name is CodeAI, trained by Terminus."},{"role":"user","content":"钉钉群名称最长能有多少个字符 ?"}],"max_tokens":2048,"temperature":1}`))
		if err != nil {
			t.Fatal(err)
		}
		request.URL.Scheme = "http"
		request.URL.Host = "ai.localhost"
		request.Method = http.MethodPost
		request.Header.Set("Content-Type", string(httputil.ApplicationJson))
		infor := reverseproxy.NewInfor(reverseproxy.NewContext(make(map[any]any)), request)
		if err := aud.SetPrompt(context.Background(), infor); err != nil {
			t.Fatal(err)
		}
		t.Logf("aud.Audit.Prompt: %s", aud.Audit.Prompt)
		if aud.Audit.Prompt != "钉钉群名称最长能有多少个字符 ?" {
			t.Fatalf("aud.Audit.Prompt error, expected: %s, got: %s", "钉钉群名称最长能有多少个字符 ?", aud.Audit.Prompt)
		}
	})
}
