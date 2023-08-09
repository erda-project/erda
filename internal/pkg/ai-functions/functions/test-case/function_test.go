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

package test_case_test

import (
	"encoding/json"
	"sigs.k8s.io/yaml"
	"testing"

	test_case "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
)

func TestSchema(t *testing.T) {
	schema := test_case.Schema
	t.Logf("json valid: %v", json.Valid(schema))
	if err := yaml.Unmarshal(schema, &schema); err != nil {
		t.Fatal(err)
	}
	t.Logf("convert to json: %s", string(schema))
	t.Logf("json valid: %v", json.Valid(schema))
}
