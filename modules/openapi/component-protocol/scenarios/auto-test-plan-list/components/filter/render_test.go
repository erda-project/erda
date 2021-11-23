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

package filter

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func TestAutoTestPlanFilter_Render(t *testing.T) {
	bdl := bundle.New()
	ctx := context.WithValue(context.Background(), protocol.GlobalInnerKeyCtxBundle.String(), protocol.ContextBundle{
		Bdl:      bdl,
		InParams: map[string]interface{}{"projectId": 25},
		Identity: apistructs.Identity{UserID: "2", OrgID: "1"},
	})

	type args struct {
		ctx      context.Context
		c        *apistructs.Component
		scenario apistructs.ComponentProtocolScenario
		event    apistructs.ComponentEvent
		gs       *apistructs.GlobalStateData
	}
	tests := []struct {
		name    string
		tpm     *AutoTestPlanFilter
		args    args
		wantErr bool
	}{
		{
			name: "filter by archive test",
			tpm:  &AutoTestPlanFilter{},
			args: args{
				event: apistructs.ComponentEvent{
					Operation: "filter",
				},
				c: &apistructs.Component{
					State: map[string]interface{}{
						"values": Value{
							Archive: []string{"archived"},
						},
					},
				},
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "initial test",
			tpm:  &AutoTestPlanFilter{},
			args: args{
				event: apistructs.ComponentEvent{
					Operation: "initial",
				},
				c: &apistructs.Component{
					State: map[string]interface{}{},
				},
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "filter by all test",
			tpm:  &AutoTestPlanFilter{},
			args: args{
				event: apistructs.ComponentEvent{
					Operation: "filter",
				},
				c: &apistructs.Component{
					State: map[string]interface{}{
						"values": Value{
							Archive: []string{"archived", "inprogress"},
						},
					},
				},
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	expects := []interface{}{true, false, nil}
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListProjectIterations", func(*bundle.Bundle, apistructs.IterationPagingRequest, string) ([]apistructs.Iteration, error) {
		return nil, nil
	})
	defer monkey.UnpatchAll()
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpm := &AutoTestPlanFilter{}
			err := tpm.Render(tt.args.ctx, tt.args.c, tt.args.scenario, tt.args.event, tt.args.gs)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoTestPlanFilter.Render() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, expects[i], tt.args.c.State["archive"])
		})
	}
}
