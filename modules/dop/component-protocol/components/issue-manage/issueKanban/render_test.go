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
//	//获取 [今天，明天，两天，七天，一个月，未来] 的时间戳
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
//			// 全选
//			reqTime: []int64{0, 0},
//			assert: []tableAssert{
//				{0, 0}, // 未定义
//				{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()},                 // 已过期
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // 今天
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // 明天
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7天
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30天
//				{thirtyDay.Unix(), 0}, // 未来
//			},
//		},
//		{
//			// 今天 - 全选
//			reqTime: []int64{nowTime.Unix(), 0},
//			assert: []tableAssert{
//				{nowTime.Unix(), 0}, // 未定义
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},    // 已过期
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // 今天
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // 明天
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7天
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30天
//				{thirtyDay.Unix(), 0}, // 未来
//			},
//		},
//		{
//			// 明天 - 7天
//			reqTime: []int64{tomorrow.Unix(), sevenDay.Unix()}, assert: []tableAssert{
//				{tomorrow.Unix(), sevenDay.Unix()},                                      // 未定义
//				{tomorrow.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // 已过期
//				{tomorrow.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // 今天
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // 明天
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},   // 7天
//				{sevenDay.Unix(), sevenDay.Unix()},                                      // 30天
//				{thirtyDay.Unix(), sevenDay.Unix()},                                     // 未来
//			},
//		},
//		{
//			// 3天 - 8天
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},  // 未定义
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // 已过期
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // 今天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // 明天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7天
//				{sevenDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},                                      // 30天
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},                                     // 未来
//			},
//		},
//		{
//			// 全选 - 今天
//			reqTime: []int64{0, nowTime.Unix()},
//			assert: []tableAssert{
//				{0, nowTime.Unix()}, // 未定义
//				{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()}, // 已过期
//				{nowTime.Unix(), nowTime.Unix()},                         // 今天
//				{tomorrow.Unix(), nowTime.Unix()},                        // 明天
//				{twoDay.Unix(), nowTime.Unix()},                          // 7天
//				{sevenDay.Unix(), nowTime.Unix()},                        // 30天
//				{thirtyDay.Unix(), nowTime.Unix()},                       // 未来
//			},
//		},
//		{
//			// 今天 - 8天
//			reqTime: []int64{nowTime.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},   // 未定义
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},   // 已过期
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},  // 今天
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // 明天
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},   // 七天                                    // 7天
//				{sevenDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()},  // 30天
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*8)).Unix()}, // 未来
//			},
//		},
//		{
//			// 3天 - 30天
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), thirtyDay.Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), thirtyDay.Unix()},                                     // 未定义
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // 已过期
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // 今天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // 明天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7天
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()},                                    // 30天
//				{thirtyDay.Unix(), thirtyDay.Unix()}, // 未来
//			},
//		},
//		{
//			// 今天 - 30天
//			reqTime: []int64{nowTime.Unix(), thirtyDay.Unix()},
//			assert: []tableAssert{
//				{nowTime.Unix(), thirtyDay.Unix()},                                       // 未定义
//				{nowTime.Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},    // 已过期
//				{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // 今天
//				{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // 明天
//				{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // 7天
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // 30天
//				{thirtyDay.Unix(), thirtyDay.Unix()},                                     // 未来
//			},
//		},
//		{
//			// 3天 - 50天
//			reqTime: []int64{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()},
//			assert: []tableAssert{
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()}, // 未定义
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), nowTime.Add(time.Second * time.Duration(-1)).Unix()},  // 已过期
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()}, // 今天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},   // 明天
//				{nowTime.Add(time.Hour * time.Duration(24*3)).Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()}, // 7天
//				{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()},                                    // 30天
//				{thirtyDay.Unix(), nowTime.Add(time.Hour * time.Duration(24*50)).Unix()},                                    // 未来
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
