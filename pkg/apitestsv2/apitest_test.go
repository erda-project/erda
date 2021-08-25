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

package apitestsv2

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/encoding/jsonpath"
)

func TestJsonPath(t *testing.T) {
	var a map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(`{"success":true}`), &a))
	data, err := jsonpath.Get(a, "success")
	assert.NoError(t, err)
	spew.Dump(data)
}
