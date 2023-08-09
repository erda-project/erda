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

package sdk_test

import (
	"encoding/json"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"testing"
)

var testCaseSchema = `
type: object
properties:
  name:
    type: string
    description: the name of the case
  preCondition:
    type: string
    description: pre condition
    items:
      type: string
  stepAndResult:
    type: array
    description: list of step and corresponding result
    items:
      type: object
      properties:
        step:
          type: string
          description: Operation steps
        result:
          type: string
          description: Expected result
required:
  - name
  - preCondition
  - stepAndResult
`

func TestFunctionDefinition_VerifyJSON(t *testing.T) {
	var df = sdk.FunctionDefinition{
		Name:        "create-test-case",
		Description: "create test case",
		Parameters:  json.RawMessage(testCaseSchema),
	}
	var j = `{"name": "dspo"}`
	if err := df.VerifyArguments(json.RawMessage(j)); err == nil {
		t.Error("err must not be nil")
	}
}
