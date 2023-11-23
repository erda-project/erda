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

package cr

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestFC(t *testing.T) {
	fc := openai.FunctionDefinition{
		Name:        "create-cr-note",
		Description: "create code review note",
		Parameters: &jsonschema.Definition{
			Type:        jsonschema.Object,
			Description: "create code review note for each code snippet inside a file",
			Properties: map[string]jsonschema.Definition{
				"fileReviewResult": {
					Type:        jsonschema.Array,
					Description: "review result for each code snippet inside a file",
					Items: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"riskScore": {
								Type:        jsonschema.Number,
								Description: "risk score",
							},
							"details": {
								Type:        jsonschema.String,
							},
						},
						Required: []string{"riskScore", "details"},
					},
					Required: []string{"fileReviewResult"},
				},
			},
			Items: &jsonschema.Definition{},
		},
	}
	b, _ := json.Marshal(fc)
	fmt.Println(string(b))
}
