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

package table

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func TestTestPlanManageTable_Render(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        *apistructs.Component
		scenario apistructs.ComponentProtocolScenario
		event    apistructs.ComponentEvent
		gs       *apistructs.GlobalStateData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_default_render",
			args: args{
				ctx:      context.Background(),
				c:        &apistructs.Component{},
				scenario: apistructs.ComponentProtocolScenario{},
				event: apistructs.ComponentEvent{
					Operation: apistructs.InitializeOperation,
				},
				gs: &apistructs.GlobalStateData{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bdl = protocol.ContextBundle{}
			var dl = bundle.Bundle{}
			monkey.PatchInstanceMethod(reflect.TypeOf(&dl), "PagingTestPlansV2", func(b *bundle.Bundle, req apistructs.TestPlanV2PagingRequest) (*apistructs.TestPlanV2PagingResponseData, error) {
				return &apistructs.TestPlanV2PagingResponseData{}, nil
			})
			bdl.Bdl = &dl
			bdl.InParams = map[string]interface{}{"projectId": 1}
			tt.args.ctx = context.WithValue(tt.args.ctx, protocol.GlobalInnerKeyCtxBundle.String(), bdl)

			tpmt := &TestPlanManageTable{}
			if err := tpmt.Render(tt.args.ctx, tt.args.c, tt.args.scenario, tt.args.event, tt.args.gs); (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
