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

package servicegroup

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/jsonstore"
)

func Test_Delete(t *testing.T) {
	js, err := jsonstore.New(jsonstore.UseMemStore())
	assert.NoError(t, err)
	sgi := ServiceGroupImpl{
		Js: js,
	}
	assert.NoError(t, sgi.Delete("services", "service-1", true, map[string]string{
		"HelloKey": "HelloValue",
	}))
}
