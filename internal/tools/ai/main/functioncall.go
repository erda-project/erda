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

package main

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	fieldName           = "name"
	fieldPreCondition   = "preCondition"
	fieldStepAndResults = "stepAndResults"
)

var createTestCase = openai.FunctionDefinition{
	Name:        "create-test-case",
	Description: "create test case",
	Parameters: &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			fieldName: {
				Type:        jsonschema.String,
				Description: "the name of test case",
			},
			fieldPreCondition: {
				Type:        jsonschema.String,
				Description: "pre condition",
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
			},
			fieldStepAndResults: {
				Type:        jsonschema.Array,
				Description: "list of step and corresponding result",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"step": {
							Type:        jsonschema.String,
							Description: "Operation steps",
						},
						"result": {
							Type:        jsonschema.String,
							Description: "Expected result",
						},
					},
				},
			},
		},
		Required: []string{fieldName, fieldPreCondition, fieldStepAndResults},
	},
}
