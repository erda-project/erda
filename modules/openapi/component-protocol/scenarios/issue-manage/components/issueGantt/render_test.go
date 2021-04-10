// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package issueGantt

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/i18n"
	"github.com/erda-project/erda/pkg/httpclient"
)

func rend(req *apistructs.ComponentProtocolRequest) (cont *apistructs.ComponentProtocolRequest, err error) {
	// bundle
	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithCMDB(),
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
//			Operation: "",
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
