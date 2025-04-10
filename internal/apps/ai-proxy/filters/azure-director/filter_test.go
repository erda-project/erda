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

package azure_director_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"

	azure_director "github.com/erda-project/erda/internal/apps/ai-proxy/filters/azure-director"
)

func TestAzureDirector_Processors(t *testing.T) {
	processors := new(azure_director.AzureDirector).AllDirectors()
	for key, v := range processors {
		t.Logf("%s: %v\n", key, v != nil)
	}
}

//func TestParseProcessorNameArgs(t *testing.T) {
//	var cases = []struct {
//		Processor string
//		Name      string
//		Args      string
//	}{
//		{
//			Processor: "SetAuthorizationIfNotSpecified",
//			Name:      "SetAuthorizationIfNotSpecified",
//		}, {
//			Processor: "TransAuthorization",
//			Name:      "TransAuthorization",
//		}, {
//			Processor: "SetAPIKeyIfNotSpecified",
//			Name:      "SetAPIKeyIfNotSpecified",
//		}, {
//			Processor: `ReplaceURIPath("/openai/deployments/${ provider.metadata.DEVELOPMENT_NAME }/completions")`,
//			Name:      "ReplaceURIPath",
//			Args:      `"/openai/deployments/${ provider.metadata.DEVELOPMENT_NAME }/completions"`,
//		},
//	}
//	for _, c := range cases {
//		name, args, err := protocol_translator.ParseProcessorNameArgs(c.Processor)
//		if err != nil {
//			t.Fatalf("failed to ParseProcessorNameArgs, err: %v", err)
//		}
//		if name != c.Name {
//			t.Fatalf("name error, expected: %s, got: %s", c.Name, name)
//		}
//		if args != c.Args {
//			t.Fatalf("args error, expected: %s, got: %s", c.Args, args)
//		}
//		t.Logf("name: %s, args: %s", name, args)
//	}
//}
//
//func TestProtocolTranslator_SetAuthorizationIfNotSpecified(t *testing.T) {
//	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewBufferString("mock body"))
//	if err != nil {
//		t.Fatal(err)
//	}
//	_ = request
//}

func TestReq(t *testing.T) {
	reqBytes, _ := os.ReadFile("req_test.json")
	tests := []struct {
		Req     openai.ChatCompletionRequest `json:"req"`
		WantErr bool
	}{
		{
			Req: openai.ChatCompletionRequest{
				ResponseFormat: &openai.ChatCompletionResponseFormat{
					JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
						Schema: &jsonschema.Definition{},
					},
				},
			},
			WantErr: false,
		},
		{
			Req:     openai.ChatCompletionRequest{},
			WantErr: true,
		},
	}

	for _, tt := range tests {
		err := json.Unmarshal(reqBytes, &tt.Req)
		if (err != nil) != tt.WantErr {
			t.Errorf("On() error = %v, wantErr %v", err, tt.WantErr)
			return
		}
	}
}
