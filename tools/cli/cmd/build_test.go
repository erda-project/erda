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

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
)

type buildRoundTripFunc func(*http.Request) (*http.Response, error)

func (f buildRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNormalizePipelineYmlName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain root pipeline", in: "pipeline.yml", want: "pipeline.yml"},
		{name: "dot slash root pipeline", in: "./pipeline.yml", want: "pipeline.yml"},
		{name: "nested pipeline", in: ".erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
		{name: "dot slash nested pipeline", in: "./.erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePipelineYmlName(tt.in); got != tt.want {
				t.Fatalf("normalizePipelineYmlName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPipelineCreateRequestCompatibility(t *testing.T) {
	request := pipelineCreateRequest(7, "main", "./pipeline.yml", false)
	if !request.AutoRun || request.AppID != 7 || request.Branch != "main" || request.PipelineYmlName != "pipeline.yml" {
		t.Fatalf("default create request = %#v", request)
	}
	parameterized := pipelineCreateRequest(7, "main", "./pipeline.yml", true)
	if parameterized.AutoRun {
		t.Fatal("parameterized create request must set autoRun=false")
	}
}

func TestParsePipelineParams(t *testing.T) {
	params, err := parsePipelineParams([]string{"verificationHoldAfterPass=60m", "other=value=with=equals"})
	if err != nil {
		t.Fatalf("parsePipelineParams() error = %v", err)
	}
	if len(params) != 2 || params[0].Name != "verificationHoldAfterPass" || params[0].Value != "60m" || params[1].Value != "value=with=equals" {
		t.Fatalf("parsePipelineParams() = %#v", params)
	}
	emptyValue, err := parsePipelineParams([]string{"empty="})
	if err != nil || len(emptyValue) != 1 || emptyValue[0].Value != "" {
		t.Fatalf("parsePipelineParams(empty value) = %#v, %v; want an empty string value", emptyValue, err)
	}

	for _, input := range [][]string{{"missing-equals"}, {"=value"}, {"name=value", "name=other"}} {
		if _, err := parsePipelineParams(input); err == nil {
			t.Errorf("parsePipelineParams(%q) expected error", input)
		}
	}
	if params, err := parsePipelineParams(nil); err != nil || params != nil {
		t.Fatalf("parsePipelineParams(nil) = %#v, %v; want nil, nil", params, err)
	}
}

func TestRunPipelineWithParamsPreservesPipelineIDOnFailure(t *testing.T) {
	client := httpclient.New()
	client.BackendClient().Transport = buildRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request", Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"success":false,"err":{"code":"HOLD_REJECTED","msg":"invalid hold"}}`))}, nil
	})
	ctx := &command.Context{CurrentHost: "http://erda.test", HttpClient: client}
	err := runPipelineWithParams(ctx, 99, []apistructs.PipelineRunParam{{Name: "verificationHoldAfterPass", Value: "60m"}})
	if err == nil || !strings.Contains(err.Error(), "pipeline 99") || !strings.Contains(err.Error(), "HOLD_REJECTED") {
		t.Fatalf("runPipelineWithParams() error = %v, want pipeline ID and server error", err)
	}
}

func TestWatchCreatedPipelineUsesSamePipelineID(t *testing.T) {
	original := pipelineStatusForRun
	t.Cleanup(func() { pipelineStatusForRun = original })
	var gotBranch string
	var gotID uint64
	var gotWatch bool
	pipelineStatusForRun = func(_ *command.Context, branch string, id uint64, watch bool) error {
		gotBranch, gotID, gotWatch = branch, id, watch
		return nil
	}
	if err := watchCreatedPipeline(&command.Context{}, "feature/hold", 1234, true); err != nil {
		t.Fatal(err)
	}
	if gotBranch != "feature/hold" || gotID != 1234 || !gotWatch {
		t.Fatalf("watch args = (%q, %d, %v), want original pipeline", gotBranch, gotID, gotWatch)
	}
}

func TestRunPipelineWithParamsUsesAuthenticatedContextClientAndContractBody(t *testing.T) {
	var gotAuth string
	var gotBody map[string]interface{}
	client := httpclient.New()
	client.BackendClient().Transport = buildRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/cicds/42/actions/run" {
			t.Fatalf("request path = %q", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"success":true}`))}, nil
	})
	ctx := &command.Context{CurrentHost: "http://erda.test", Token: "token-value", HttpClient: client}
	if err := runPipelineWithParams(ctx, 42, []apistructs.PipelineRunParam{{Name: "verificationHoldAfterPass", Value: "60m"}}); err != nil {
		t.Fatalf("runPipelineWithParams() error = %v", err)
	}
	if gotAuth != "token-value" {
		t.Fatalf("Authorization = %q, want context token", gotAuth)
	}
	params, ok := gotBody["pipelineRunParams"].([]interface{})
	if !ok || len(params) != 1 || params[0].(map[string]interface{})["name"] != "verificationHoldAfterPass" || params[0].(map[string]interface{})["value"] != "60m" {
		t.Fatalf("request body = %#v, want pipelineRunParams contract", gotBody)
	}
}
