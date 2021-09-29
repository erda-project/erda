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

package apiEditor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenEmptyAPISpecStr(t *testing.T) {
	testEmptyAPISpec, testEmptyAPISpecStr := genEmptyAPISpecStr()
	assert.Equal(t, "GET", testEmptyAPISpec.APIInfo.Method)
	assert.Equal(t, `{"apiSpec":{"id":"","name":"","url":"","method":"GET","headers":null,"params":null,"body":{"type":"","content":null},"out_params":null,"asserts":null},"loop":null}`,
		testEmptyAPISpecStr)
}
