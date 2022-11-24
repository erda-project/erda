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

func TestFindValue(t *testing.T) {
	request, _ := http.NewRequest("GET", "localhost:9529/api/dashboard/blocks?ttt=123", nil)
	service := queryValue{}

	tests := []struct {
		name string
		expr string
		want interface{}
	}{
		{
			name: "normal",
			expr: "ttt",
			want: "123",
		}, {
			name: "no_expr",
			want: "",
		}, {
			name: "no_data",
			expr: "aabbccc",
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, service.get(test.expr, request))
		})
	}
}
