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

package command

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/status"
)

func TestFetchOpenapiDoesNotPrintInfoByDefault(t *testing.T) {
	ctx, out := newFetchOpenapiTestContext(t, false)

	require.NoError(t, ctx.FetchOpenapi())
	require.Empty(t, out.String())
}

func TestFetchOpenapiPrintsInfoInDebugMode(t *testing.T) {
	ctx, out := newFetchOpenapiTestContext(t, true)

	require.NoError(t, ctx.FetchOpenapi())
	require.Contains(t, out.String(), "erda openapi info")
	require.Contains(t, out.String(), "dice_version: 2.4")
}

func newFetchOpenapiTestContext(t *testing.T, debug bool) (*Context, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer
	origOutput := contextOutput
	contextOutput = func() io.Writer {
		return &out
	}
	t.Cleanup(func() {
		contextOutput = origOutput
	})

	client := httpclient.New()
	client.BackendClient().Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Equal(t, "/metadata.json", r.URL.Path)
		body := fmt.Sprintf(`{
			"openapi_public_url": "%s",
			"version": {
				"built": "2026-03-23 15:24:01",
				"dice_version": "2.4",
				"git_commit": "301f210588f3472eae57e654d528037aa1facedc",
				"go_version": "go version go1.24.6 linux/amd64"
			}
		}`, serverURLWithOpenapiHost(r))
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	return &Context{
		CurrentHost: "http://127.0.0.1:12345",
		Debug:       debug,
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}, &out
}

func serverURLWithOpenapiHost(r *http.Request) string {
	return "http://openapi." + r.Host
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
