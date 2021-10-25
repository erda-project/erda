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

package issueGantt

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-manage/issueGantt/gantt"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/i18n"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func rend(req *apistructs.ComponentProtocolRequest) (cont *apistructs.ComponentProtocolRequest, err error) {
	// bundle
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithCoreServices(),
	}
	bdl := bundle.New(bundleOpts...)

	r := http.Request{}
	i18nPrinter := i18n.I18nPrinter(&r)
	inParams := req.InParams
	identity := apistructs.Identity{UserID: "2", OrgID: "1"}
	ctxBdl := protocol.ContextBundle{
		Bdl:         bdl,
		I18nPrinter: i18nPrinter,
		InParams:    inParams,
		Identity:    identity,
	}
	ctx := context.Background()
	ctx1 := context.WithValue(ctx, protocol.GlobalInnerKeyCtxBundle.String(), ctxBdl)

	err = protocol.RunScenarioRender(ctx1, req)
	if err != nil {
		return
	}
	cont = req
	return
}

//func TestRender(t *testing.T) {
//	req := apistructs.ComponentProtocolRequest{
//		Scenario: apistructs.ComponentProtocolScenario{
//			ScenarioKey:  "issueGantt",
//			ScenarioType: "issue-manage",
//		},
//		Event: apistructs.ComponentEvent{
//			Component: "",
//			GenerateOperation: "",
//		},
//		InParams: map[string]interface{}{"projectId": "11"},
//	}
//	content, err := rend(&req)
//	ctxByte, err := json.Marshal(*content)
//	if err != nil {
//		t.Errorf("marshal error:%v", err)
//		return
//	}
//	t.Logf("marshal content:%s", string(ctxByte))
//}

//func TestFilter(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	bundleOpts := []bundle.Option{
//		bundle.WithHTTPClient(
//			httpclient.New(
//				httpclient.WithTimeout(time.Second, time.Second*60),
//			)),
//		bundle.WithCMDB(),
//	}
//	bdl := bundle.New(bundleOpts...)
//	comp := Gantt{
//		CtxBdl: protocol.ContextBundle{
//			Bdl:         bdl,
//			I18nPrinter: i18n.I18nPrinter(nil),
//			Identity:    apistructs.Identity{},
//			InParams:    nil,
//		},
//	}
//	req := IssueFilterRequest{
//		IssuePagingRequest: apistructs.IssuePagingRequest{
//			PageSize: 200,
//			IssueListRequest: apistructs.IssueListRequest{
//				Title:       "",
//				Type:        []apistructs.IssueType{apistructs.IssueTypeTask},
//				ProjectID:   35,
//				IterationID: 0,
//				IdentityInfo: apistructs.IdentityInfo{
//					UserID: "2",
//				},
//				External: false,
//			},
//		},
//		BoardType: BoardTypeCustom,
//	}
//	ib, err := comp.Filter(req)
//	if err != nil {
//		panic(err)
//	}
//	t.Logf("issue board:%v", ib)
//}

// genData() results are the same every time
func TestGenData(t *testing.T) {
	p := Gantt{}
	issues := []apistructs.Issue{
		{
			ID:       1,
			Assignee: "1",
		},
		{
			ID:       2,
			Assignee: "1",
		},
		{
			ID:       3,
			Assignee: "1",
		},
		{
			ID:       4,
			Assignee: "2",
		},
		{
			ID:       5,
			Assignee: "2",
		},
		{
			ID:       6,
			Assignee: "3",
		},
		{
			ID:       6,
			Assignee: "3",
		},
	}
	t1 := time.Now()
	t2 := time.Now()
	res := map[int]interface{}{}
	for i := 0; i < 10; i++ {
		p.Data = gantt.Data{}
		err := p.genData(issues, &t1, &t2)
		assert.NoError(t, err)
		res[i] = p.Data
		if i > 0 {
			assert.Equal(t, res[i-1], res[i])
		}
	}
}

func TestSetStateToUrlQuery(t *testing.T) {
	state := gantt.State{
		Total:               11,
		PageNo:              1,
		PageSize:            10,
		IssueViewGroupValue: "foo",
		IssueType:           "bar",
	}
	var g Gantt
	g.State = state
	c := &cptype.Component{
		State: make(map[string]interface{}),
	}
	if err := g.SetStateToUrlQuery(c); err != nil {
		t.Error("fail")
	}
	// {"pageNo":1,"pageSize":10,"issueViewGroupValue":"foo","IssueType":"bar"}
	if c.State["issueGantt__urlQuery"] != "eyJwYWdlTm8iOjEsInBhZ2VTaXplIjoxMCwiaXNzdWVWaWV3R3JvdXBWYWx1ZSI6ImZvbyIsIklzc3VlVHlwZSI6ImJhciJ9" {
		t.Error("fail")
	}
}

func TestGantt_Export(t *testing.T) {
	type fields struct {
		sdk             *cptype.SDK
		bdl             *bundle.Bundle
		Uids            []string
		CommonGantt     gantt.CommonGantt
		DefaultProvider base.DefaultProvider
	}
	type args struct {
		c  *cptype.Component
		gs *cptype.GlobalStateData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "with default value",
			args: args{
				c: &cptype.Component{
					Data: map[string]interface{}{
						"list": []gantt.DataItem{
							{
								ID: 2,
							},
						},
					},
				},
				gs: &cptype.GlobalStateData{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gantt{
				sdk:             tt.fields.sdk,
				bdl:             tt.fields.bdl,
				Uids:            tt.fields.Uids,
				CommonGantt:     tt.fields.CommonGantt,
				DefaultProvider: tt.fields.DefaultProvider,
			}
			if err := g.Export(tt.args.c, tt.args.gs); (err != nil) != tt.wantErr {
				t.Errorf("Gantt.Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, nil, tt.args.c.Data["list"])
		})
	}
}
