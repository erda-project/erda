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

package autotest_cookie_keep_before

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/pipeline/report/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/internal/tools/pipeline/aop/plugins/task/autotest_cookie_keep_after"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/report"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/apitestsv2"
)

func Test_appendOrReplaceSetCookiesToCookie(t *testing.T) {
	type args struct {
		setCookies   []string
		originCookie string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "all empty",
			args: args{
				setCookies:   nil,
				originCookie: "",
			},
			want: "",
		},
		{
			name: "set-cookie of one existed cookie",
			args: args{
				setCookies:   []string{`cookie_a=aa; Path=/; Domain=xxx; Expires=Fri, 03 Sep 2021 15:12:15 GMT`},
				originCookie: "cookie_a=a; cookie_b=b",
			},
			want: "cookie_a=aa; cookie_b=b",
		},
		{
			name: "set-cookie of tow existed cookie(a,c) and append one(d), and originCookie is splitted by `;`, not `; `",
			args: args{
				setCookies:   []string{`cookie_a=aa; Path=/; Domain=xxx; Expires=Fri, 03 Sep 2021 15:12:15 GMT`, `cookie_c=C`, `cookie_d=dd`},
				originCookie: "cookie_a=a;cookie_b=b;cookie_c=c",
			},
			want: "cookie_a=aa; cookie_b=b; cookie_c=C; cookie_d=dd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendOrReplaceSetCookiesToCookie(tt.args.setCookies, tt.args.originCookie); got != tt.want {
				t.Errorf("appendOrReplaceSetCookiesToCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	db := &dbclient.Client{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineTaskExtra", func(_ *dbclient.Client, id uint64, extra spec.PipelineTaskExtra, ops ...dbclient.SessionOption) error {
		return nil
	})
	defer pm1.Unpatch()
	mockReport := &report.MockReport{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(mockReport), "QueryPipelineReportSet", func(_ *report.MockReport, _ context.Context, _ *pb.PipelineReportSetQueryRequest) (*pb.PipelineReportSetQueryResponse, error) {
		cookieMeta1 := map[string]interface{}{
			"Set-Cookie": `[
    "BAIDUID=EA64A26D7004088DC84439A66DB1EC6E:FG=1; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com",
    "BIDUPSID=EA64A26D7004088DC84439A66DB1EC6E; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com",
    "PSTM=1657609331; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com",
    "BAIDUID=EA64A26D7004088D02646E5A79CBAF9D:FG=1; max-age=31536000; expires=Wed, 12-Jul-23 07:02:11 GMT; domain=.baidu.com; path=/; version=1; comment=bd",
    "BDSVRTM=0; path=/",
    "BD_HOME=1; path=/",
    "H_PS_PSSID=36548_36463_36726_36454_36452_36691_36166_36695_36698_36816_36569_36530_36772_36730_36746_36761_36768_36764_26350_36649; path=/; domain=.baidu.com"
]`,
		}
		cookieMeta2 := map[string]interface{}{
			"Set-Cookie": `[
    "BDSVRTM=0; path=/",
    "BD_HOME=1; path=/",
    "H_PS_PSSID=36559_36750_36726_36454_36453_36692_36167_36695_36696_36816_36570_36530_36772_36746_36762_36768_36766_26350; path=/; domain=.baidu.com"
]`,
		}
		pbMeta1, _ := structpb.NewStruct(cookieMeta1)
		pbMeta2, _ := structpb.NewStruct(cookieMeta2)
		return &pb.PipelineReportSetQueryResponse{
			Data: &pb.PipelineReportSet{
				Reports: []*pb.PipelineReport{
					{
						Meta: pbMeta1,
					},
					{
						Meta: pbMeta2,
					},
				},
			},
		}, nil
	})
	defer pm2.Unpatch()
	ctx := &aoptypes.TuneContext{
		SDK: aoptypes.SDK{
			Task: spec.PipelineTask{
				Extra: spec.PipelineTaskExtra{
					PrivateEnvs: map[string]string{
						autotest_cookie_keep_after.AutotestApiGlobalConfig: "{}",
					},
				},
				Type: taskType,
			},
			Pipeline: spec.Pipeline{},
			DBClient: db,
			Report:   mockReport,
		},
	}
	p := provider{}
	err := p.Handle(ctx)
	assert.NoError(t, err)
	var config apistructs.AutoTestAPIConfig
	err = json.Unmarshal([]byte(ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig]), &config)
	assert.NoError(t, err)
	cookie := config.Header[apitestsv2.HeaderCookie]
	assert.Equal(t, "BAIDUID=EA64A26D7004088D02646E5A79CBAF9D:FG=1; BIDUPSID=EA64A26D7004088DC84439A66DB1EC6E; PSTM=1657609331; BDSVRTM=0; BD_HOME=1; H_PS_PSSID=36559_36750_36726_36454_36453_36692_36167_36695_36696_36816_36570_36530_36772_36746_36762_36768_36766_26350", cookie)
	cookieLst := strings.Split(cookie, "; ")
	assert.Equal(t, 6, len(cookieLst))
	assert.Equal(t, "H_PS_PSSID=36559_36750_36726_36454_36453_36692_36167_36695_36696_36816_36570_36530_36772_36746_36762_36768_36766_26350", cookieLst[5])
	assert.Equal(t, "BIDUPSID=EA64A26D7004088DC84439A66DB1EC6E", cookieLst[1])
}

func Test_getSortedReports(t *testing.T) {
	type args struct {
		reportSet *pb.PipelineReportSetQueryResponse
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "desc reports",
			args: args{
				reportSet: &pb.PipelineReportSetQueryResponse{
					Data: &pb.PipelineReportSet{
						Reports: []*pb.PipelineReport{
							{
								ID:         1003,
								PipelineID: 1,
							},
							{
								ID:         1002,
								PipelineID: 1,
							},
							{
								ID:         1001,
								PipelineID: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "asc reports",
			args: args{
				reportSet: &pb.PipelineReportSetQueryResponse{
					Data: &pb.PipelineReportSet{
						Reports: []*pb.PipelineReport{
							{
								ID:         1001,
								PipelineID: 1,
							},
							{
								ID:         1002,
								PipelineID: 1,
							},
							{
								ID:         1003,
								PipelineID: 1,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		res := getSortedReports(tt.args.reportSet)
		for i := 1; i < len(res); i++ {
			assert.True(t, res[i].ID > res[i-1].ID)
		}
	}
}
