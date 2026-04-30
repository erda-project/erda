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

func TestBuildBranchWorkspaceQueryParams(t *testing.T) {
	params := buildBranchWorkspaceQueryParams(3001)

	if got := params["appID"]; got != "3001" {
		t.Fatalf("appID = %q, want 3001", got)
	}
	if _, ok := params["appId"]; ok {
		t.Fatal("appId should not be used; backend expects appID")
	}
}

func TestInspectRuntimeIncludesOrgIDHeader(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/runtimes/1008165" {
			t.Fatalf("request path = %q, want /api/runtimes/1008165", got)
		}
		if got := r.URL.Query().Get("applicationId"); got != "1003418" {
			t.Fatalf("applicationId = %q, want 1003418", got)
		}
		if got := r.URL.Query().Get("workspace"); got != "PROD" {
			t.Fatalf("workspace = %q, want PROD", got)
		}
		if got := r.Header.Get("Org-ID"); got != "141" {
			t.Fatalf("Org-ID header = %q, want 141", got)
		}
		if got := r.Header.Get("org"); got != "141" {
			t.Fatalf("org header = %q, want 141", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":{"id":1008165},"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	runtime, err := InspectRuntime(ctx, 141, 1008165, 1003418, "PROD")
	if err != nil {
		t.Fatalf("InspectRuntime() error = %v", err)
	}
	if runtime.ID != 1008165 {
		t.Fatalf("runtime.ID = %d, want 1008165", runtime.ID)
	}
}

func TestListRuntimeServicePodsIncludesOrgIDHeader(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/instances/actions/get-service-pods" {
			t.Fatalf("request path = %q, want /api/instances/actions/get-service-pods", got)
		}
		if got := r.URL.Query().Get("runtimeID"); got != "1008165" {
			t.Fatalf("runtimeID = %q, want 1008165", got)
		}
		if got := r.URL.Query().Get("serviceName"); got != "web" {
			t.Fatalf("serviceName = %q, want web", got)
		}
		if got := r.Header.Get("Org-ID"); got != "141" {
			t.Fatalf("Org-ID header = %q, want 141", got)
		}
		if got := r.Header.Get("org"); got != "141" {
			t.Fatalf("org header = %q, want 141", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":[],"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	if _, err := ListRuntimeServicePods(ctx, 141, 1008165, "web"); err != nil {
		t.Fatalf("ListRuntimeServicePods() error = %v", err)
	}
}

func TestListRuntimeServiceStoppedContainersIncludesOrgIDHeaderAndStatus(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/instances/actions/get-service" {
			t.Fatalf("request path = %q, want /api/instances/actions/get-service", got)
		}
		if got := r.URL.Query().Get("runtimeID"); got != "1008165" {
			t.Fatalf("runtimeID = %q, want 1008165", got)
		}
		if got := r.URL.Query().Get("serviceName"); got != "web" {
			t.Fatalf("serviceName = %q, want web", got)
		}
		if got := r.URL.Query().Get("status"); got != "stopped" {
			t.Fatalf("status = %q, want stopped", got)
		}
		if got := r.Header.Get("Org-ID"); got != "141" {
			t.Fatalf("Org-ID header = %q, want 141", got)
		}
		if got := r.Header.Get("org"); got != "141" {
			t.Fatalf("org header = %q, want 141", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":[],"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	if _, err := ListRuntimeServiceStoppedContainers(ctx, 141, 1008165, "web"); err != nil {
		t.Fatalf("ListRuntimeServiceStoppedContainers() error = %v", err)
	}
}

func TestGetRuntimePodLogsUsesRuntimeLogsPathAndPodMetadata(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/runtime/logs" {
			t.Fatalf("request path = %q, want /api/runtime/logs", got)
		}
		if got := r.URL.Query().Get("applicationId"); got != "1003418" {
			t.Fatalf("applicationId = %q, want 1003418", got)
		}
		if got := r.URL.Query().Get("id"); got != "container-1" {
			t.Fatalf("id = %q, want container-1", got)
		}
		if got := r.URL.Query().Get("clusterName"); got != "erda-cloud" {
			t.Fatalf("clusterName = %q, want erda-cloud", got)
		}
		if got := r.URL.Query().Get("podName"); got != "web-0" {
			t.Fatalf("podName = %q, want web-0", got)
		}
		if got := r.URL.Query().Get("podNamespace"); got != "project-387-prod" {
			t.Fatalf("podNamespace = %q, want project-387-prod", got)
		}
		if got := r.URL.Query().Get("containerName"); got != "web" {
			t.Fatalf("containerName = %q, want web", got)
		}
		if got := r.URL.Query().Get("live"); got != "true" {
			t.Fatalf("live = %q, want true", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":{"lines":[]},"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	pod := apistructs.Pod{
		ClusterName:  "erda-cloud",
		PodName:      "web-0",
		K8sNamespace: "project-387-prod",
	}
	if _, err := GetRuntimePodLogs(ctx, "erda", 1003418, pod, "web", "container-1", "stdout", 200); err != nil {
		t.Fatalf("GetRuntimePodLogs() error = %v", err)
	}
}

func TestGetRuntimeStoppedContainerLogsUsesLiveFalse(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/runtime/logs" {
			t.Fatalf("request path = %q, want /api/runtime/logs", got)
		}
		if got := r.URL.Query().Get("id"); got != "container-dead" {
			t.Fatalf("id = %q, want container-dead", got)
		}
		if got := r.URL.Query().Get("podName"); got != "web-dead-0" {
			t.Fatalf("podName = %q, want web-dead-0", got)
		}
		if got := r.URL.Query().Get("podNamespace"); got != "project-387-prod" {
			t.Fatalf("podNamespace = %q, want project-387-prod", got)
		}
		if got := r.URL.Query().Get("containerName"); got != "web" {
			t.Fatalf("containerName = %q, want web", got)
		}
		if got := r.URL.Query().Get("live"); got != "false" {
			t.Fatalf("live = %q, want false", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":{"lines":[]},"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	container := apistructs.Container{
		Service:     "web",
		ClusterName: "erda-cloud",
		ContainerID: "container-dead",
		K8sInstanceMetaInfo: apistructs.K8sInstanceMetaInfo{
			PodName:      "web-dead-0",
			PodNamespace: "project-387-prod",
		},
	}
	if _, err := GetRuntimeStoppedContainerLogs(ctx, "erda", 1003418, container, "", 200); err != nil {
		t.Fatalf("GetRuntimeStoppedContainerLogs() error = %v", err)
	}
}

func TestGetRuntimePodLogsWithOptionsUsesIncrementalQuery(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = runtimeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Path; got != "/api/runtime/logs" {
			t.Fatalf("request path = %q, want /api/runtime/logs", got)
		}
		if got := r.URL.Query().Get("start"); got != "1000" {
			t.Fatalf("start = %q, want 1000", got)
		}
		if got := r.URL.Query().Get("count"); got != "50" {
			t.Fatalf("count = %q, want 50", got)
		}
		if got := r.URL.Query().Get("stream"); got != "stdout" {
			t.Fatalf("stream = %q, want stdout", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"data":{"lines":[]},"err":{"code":"","msg":"","ctx":null}}`)),
		}, nil
	})

	ctx := &command.Context{
		CurrentHost: "http://127.0.0.1:12345",
		Sessions:    map[string]status.StatusInfo{},
		HttpClient:  client,
	}

	pod := apistructs.Pod{
		ClusterName:  "erda-cloud",
		PodName:      "web-0",
		K8sNamespace: "project-387-prod",
	}
	if _, err := GetRuntimePodLogsWithOptions(ctx, "erda", 1003418, pod, "web", "container-1", RuntimeLogOptions{
		Stream: "stdout",
		Start:  1000,
		Count:  50,
		Tail:   200,
	}); err != nil {
		t.Fatalf("GetRuntimePodLogsWithOptions() error = %v", err)
	}
}

type runtimeRoundTripFunc func(*http.Request) (*http.Response, error)

func (f runtimeRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
