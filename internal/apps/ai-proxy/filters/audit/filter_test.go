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
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/audit"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

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
		if err := aud.SetPrompt(context.Background(), &mockedInfor{
			u: &url.URL{
				Scheme: "https",
				Host:   "ai.localhost",
				Path:   "/v1/chat/completions",
			},
			method: http.MethodPost,
			header: http.Header{
				"Content-Type": []string{string(httputil.ApplicationJson)},
			},
			body:       bytes.NewBufferString(`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"Your Name is CodeAI, trained by Terminus."},{"role":"user","content":"钉钉群名称最长能有多少个字符 ?"}],"max_tokens":2048,"temperature":1}`),
			statusCode: http.StatusOK,
			remoteAddr: "127.0.0.1",
		}); err != nil {
			t.Fatal(err)
		}
		t.Logf("aud.Audit.Prompt: %s", aud.Audit.Prompt)
		if aud.Audit.Prompt != "钉钉群名称最长能有多少个字符 ?" {
			t.Fatalf("aud.Audit.Prompt error, expected: %s, got: %s", "钉钉群名称最长能有多少个字符 ?", aud.Audit.Prompt)
		}
	})
}

var _ reverseproxy.HttpInfor = (*mockedInfor)(nil)

type mockedInfor struct {
	u          *url.URL
	method     string
	header     http.Header
	body       *bytes.Buffer
	statusCode int
	remoteAddr string
}

func (m *mockedInfor) Body() io.ReadCloser {
	return io.NopCloser(m.body)
}

func (m *mockedInfor) BodyBuffer() *bytes.Buffer {
	//TODO implement me
	panic("implement me")
}

func (m *mockedInfor) Mutex() *sync.Mutex {
	//TODO implement me
	panic("implement me")
}

func (m *mockedInfor) Method() string {
	return m.method
}

func (m *mockedInfor) URL() *url.URL {
	return m.u
}

func (m *mockedInfor) Status() string {
	return http.StatusText(m.statusCode)
}

func (m *mockedInfor) StatusCode() int {
	return m.statusCode
}

func (m *mockedInfor) Header() http.Header {
	return m.header
}

func (m *mockedInfor) ContentLength() int64 {
	return int64(len(m.body.Bytes()))
}

func (m *mockedInfor) Host() string {
	return m.u.Host
}

func (m *mockedInfor) RemoteAddr() string {
	return m.remoteAddr
}
