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

package match

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockValue struct {
}

func (m mockValue) get(expr string, r *http.Request) interface{} {
	return "test"
}

func TestGet(t *testing.T) {
	ValueFunc = make(map[string]value)

	registry("mock", mockValue{})

	tests := []struct {
		name string
		expr []string
		want map[string]interface{}
	}{
		{
			name: "normal",
			expr: []string{
				"mock:ttt",
			},
			want: map[string]interface{}{"ttt": "test"},
		},
		{
			name: "no_type",
			expr: []string{
				"ttt",
			},
			want: nil,
		},
		{
			name: "no_type",
			expr: []string{
				":ttt",
			},
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, Get(test.expr, nil))
		})
	}
}
