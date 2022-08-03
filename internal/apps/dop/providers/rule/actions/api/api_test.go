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

package api

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/internal/apps/dop/providers/rule/jsonnet"
)

func Test_provider_getAPIConfig(t *testing.T) {
	var engine *jsonnet.Engine
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(engine), "EvaluateBySnippet",
		func(d *jsonnet.Engine, snippet string, configs []jsonnet.TLACodeConfig) (string, error) {
			return "", nil
		},
	)

	defer p1.Unpatch()
	p := &provider{TemplateParser: engine}
	type args struct {
		api *API
	}
	tests := []struct {
		name    string
		args    args
		want    *APIConfig
		wantErr bool
	}{
		{
			args: args{
				api: &API{
					URL: "http://localhost:9090/api/test",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.getAPIConfig(tt.args.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.getAPIConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("provider.getAPIConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
