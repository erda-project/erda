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

package issueKanban

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
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

//func TestFilter(t *testing.T) {
//	req := apistructs.ComponentProtocolRequest{
//		Scenario: apistructs.ComponentProtocolScenario{
//			ScenarioKey:  "issueKanban",
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
//func TestTimeBoard(t *testing.T) {
//	bundleOpts := []bundle.Option{}
//	bdl := bundle.New(bundleOpts...)
//	i := ComponentIssueBoard{}
//	b := protocol.ContextBundle{
//		Bdl: bdl,
//	}
//	i.ctxBdl = b
//	i.boardType = BoardTypeTime
//
//	//?????? [??????????????????????????????????????????????????????] ????????????
//	nowTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
//	tomorrow := nowTime.Add(time.Hour * time.Duration(24))
//	twoDay := nowTime.Add(time.Hour * time.Duration(24*2))
//	sevenDay := nowTime.Add(time.Hour * time.Duration(24*7))
//	thirtyDay := nowTime.Add(time.Hour * time.Duration(24*30))
//	type tableAssert struct {
//		StartTime int64
//		EndTime   int64
//	}
//	var tables = []struct {
//		reqTime []int64 //
//		assert  []tableAssert
//	}{
//		{
//			// ??????
//			reqTime: []int64{0, 0},
//			assert: []tableAssert{
//				{0, 0}, // ?????????
//				{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()},                 // ?????????
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // ??????
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7???
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30???
//				{thirtyDay.Unix(), 0}, // ??????
//			},
//		},
//		{
//			// ?????? - ??????
//			reqTime: []int64{nowTime.Unix(), 0},
//			assert: []tableAssert{
//				{nowTime.Unix(), 0}, // ?????????
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},    // ?????????
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // ??????
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7???
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30???
//				{thirtyDay.Unix(), 0}, // ??????
//			},
//		},
//		{
//			// ?????? - 7???
//			reqTime: []int64{tomorrow.Unix(), sevenDay.Unix()}, assert: []tableAssert{
//				{tomorrow.Unix(), sevenDay.Unix()},                                      // ?????????
//				{tomorrow.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // ?????????
//				{tomorrow.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // ??????
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},   // 7???
//				{sevenDay.Unix(), sevenDay.Unix()},                                      // 30???
//				{thirtyDay.Unix(), sevenDay.Unix()},                                     // ??????
//			},
//		},
//		{
//			// 3??? - 8???
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},  // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7???
//				{sevenDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},                                      // 30???
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},                                     // ??????
//			},
//		},
//		{
//			// ?????? - ??????
//			reqTime: []int64{0, nowTime.Unix()},
//			assert: []tableAssert{
//				{0, nowTime.Unix()}, // ?????????
//				{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()}, // ?????????
//				{nowTime.Unix(), nowTime.Unix()},                         // ??????
//				{tomorrow.Unix(), nowTime.Unix()},                        // ??????
//				{twoDay.Unix(), nowTime.Unix()},                          // 7???
//				{sevenDay.Unix(), nowTime.Unix()},                        // 30???
//				{thirtyDay.Unix(), nowTime.Unix()},                       // ??????
//			},
//		},
//		{
//			// ?????? - 8???
//			reqTime: []int64{nowTime.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},   // ?????????
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},   // ?????????
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},  // ??????
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????                                    // 7???
//				{sevenDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},  // 30???
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()}, // ??????
//			},
//		},
//		{
//			// 3??? - 30???
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), thirtyDay.Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), thirtyDay.Unix()},                                     // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7???
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()},                                    // 30???
//				{thirtyDay.Unix(), thirtyDay.Unix()}, // ??????
//			},
//		},
//		{
//			// ?????? - 30???
//			reqTime: []int64{nowTime.Unix(), thirtyDay.Unix()},
//			assert: []tableAssert{
//				{nowTime.Unix(), thirtyDay.Unix()},                                       // ?????????
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},    // ?????????
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // ??????
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7???
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30???
//				{thirtyDay.Unix(), thirtyDay.Unix()},                                     // ??????
//			},
//		},
//		{
//			// 3??? - 50???
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()}, // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // ?????????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // ??????
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7???
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()},                                    // 30???
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()},                                    // ??????
//			},
//		},
//	}
//
//	for tableIndex, tmlist := range tables {
//		p := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "PageIssues", func(_ *bundle.Bundle, r apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
//			endTime := r.EndFinishedAt
//			createdAt := r.StartFinishedAt
//			return &apistructs.IssuePagingResponse{
//				Data: &apistructs.IssuePagingResponseData{
//					List: []apistructs.Issue{
//						{
//							IterationID: endTime,
//							ID:          createdAt,
//						},
//					},
//				},
//			}, nil
//		})
//		defer p.Unpatch()
//		var req apistructs.IssuePagingRequest
//		req.StartFinishedAt = tmlist.reqTime[0] * 1000
//		req.EndFinishedAt = tmlist.reqTime[1] * 1000
//		list, _, _ := i.FilterByTime(req)
//		for resultIndex, v := range list {
//			issueCert := v.List[0]
//			assert.Equal(t, tables[tableIndex].assert[resultIndex].EndTime, issueCert.IterationID/1000)
//			assert.Equal(t, tables[tableIndex].assert[resultIndex].StartTime, issueCert.ID/1000)
//		}
//	}
//}
