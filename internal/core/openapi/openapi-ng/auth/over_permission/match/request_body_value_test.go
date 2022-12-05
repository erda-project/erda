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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

func TestFormBodyGet(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		expr string
		want interface{}
	}{
		{
			name: "render scope",
			data: "{\"scenario\":{\"scenarioType\":\"msp-alert-event-list\",\"scenarioKey\":\"msp-alert-event-list\"},\"inParams\":{\"scopeId\":\"633\",\"scope\":\"org\"}}",
			expr: "inParams.scope",
			want: "org",
		},
		{
			name: "render scope id",
			data: "{\"scenario\":{\"scenarioType\":\"msp-alert-event-list\",\"scenarioKey\":\"msp-alert-event-list\"},\"inParams\":{\"scopeId\":\"633\",\"scope\":\"org\"}}",
			expr: "inParams.scopeId",
			want: "633",
		},
	}

	service := requestBody{}
	for _, test := range tests {
		request, _ := http.NewRequest("POST", "123.com", strings.NewReader(jsonparse.JsonOneLine(test.data)))
		t.Run(test.name, func(t *testing.T) {
			got := service.get(test.expr, request)
			require.Equal(t, test.want, got)
		})
	}
}
