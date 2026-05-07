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

package common

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/status"
)

func TestListMyIssueIncludesOrgIDHeader(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = issueRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/issues" {
			t.Fatalf("request path = %q, want /api/issues", got)
		}
		if got := r.Header.Get("Org-ID"); got != "1" {
			t.Fatalf("Org-ID header = %q, want 1", got)
		}
		if got := r.Header.Get("org"); got != "1" {
			t.Fatalf("org header = %q, want 1", got)
		}
		if got := r.URL.Query().Get("orgID"); got != "1" {
			t.Fatalf("orgID = %q, want 1", got)
		}
		if got := r.URL.Query().Get("projectID"); got != "2" {
			t.Fatalf("projectID = %q, want 2", got)
		}
		if got := r.URL.Query().Get("type"); got != "TASK" {
			t.Fatalf("type = %q, want TASK", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":{"total":0,"list":[]},"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	_, err := ListMyIssue(ctx, &apistructs.IssuePagingRequest{
		OrgID:    1,
		PageNo:   1,
		PageSize: 20,
		IssueListRequest: apistructs.IssueListRequest{
			ProjectID: 2,
			Type:      []apistructs.IssueType{apistructs.IssueTypeTask},
		},
	})
	if err != nil {
		t.Fatalf("ListMyIssue() error = %v", err)
	}
}

type issueRoundTripFunc func(*http.Request) (*http.Response, error)

func (f issueRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
