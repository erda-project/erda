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

package container

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/service/common/custom"
	"github.com/erda-project/erda/pkg/common/apis"
)

type mockI18n struct {
}

func (m mockI18n) Get(lang i18n.LanguageCodes, key, def string) string {
	return def
}

func (m mockI18n) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

func (m mockI18n) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

type mockMetricService struct {
	t                          *testing.T
	checkQueryWithInfluxFormat func(t *testing.T, ctx context.Context, request *metricpb.QueryWithInfluxFormatRequest)
}

func (m mockMetricService) QueryWithInfluxFormat(ctx context.Context, request *metricpb.QueryWithInfluxFormatRequest) (*metricpb.QueryWithInfluxFormatResponse, error) {
	m.checkQueryWithInfluxFormat(m.t, ctx, request)

	return &metricpb.QueryWithInfluxFormatResponse{
		Results: []*metricpb.Result{
			{
				Series: []*metricpb.Serie{
					{
						Columns: []string{""},
						Rows: []*metricpb.Row{
							{
								Values: []*structpb.Value{
									{},
									{},
									{},
									{},
									{},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m mockMetricService) SearchWithInfluxFormat(ctx context.Context, request *metricpb.QueryWithInfluxFormatRequest) (*metricpb.QueryWithInfluxFormatResponse, error) {
	return nil, nil
}

func (m mockMetricService) QueryWithTableFormat(ctx context.Context, request *metricpb.QueryWithTableFormatRequest) (*metricpb.QueryWithTableFormatResponse, error) {
	return nil, nil
}

func (m mockMetricService) SearchWithTableFormat(ctx context.Context, request *metricpb.QueryWithTableFormatRequest) (*metricpb.QueryWithTableFormatResponse, error) {
	return nil, nil
}

func (m mockMetricService) GeneralQuery(ctx context.Context, request *metricpb.GeneralQueryRequest) (*metricpb.GeneralQueryResponse, error) {
	return nil, nil
}

func (m mockMetricService) GeneralSearch(ctx context.Context, request *metricpb.GeneralQueryRequest) (*metricpb.GeneralQueryResponse, error) {
	return nil, nil
}

func TestRegisterInitializeOpContext(t *testing.T) {
	p := provider{
		I18n: mockI18n{},
		Metric: mockMetricService{
			t: t,
			checkQueryWithInfluxFormat: func(t *testing.T, ctx context.Context, request *metricpb.QueryWithInfluxFormatRequest) {
				require.NotNil(t, ctx)
				require.Equal(t, "orgname", apis.GetHeader(ctx, "org"))
				require.Equal(t, "", apis.GetHeader(ctx, "terminus_key"))
			},
		},
		ServiceInParams: custom.ServiceInParams{
			InParamsPtr: &custom.Model{
				StartTime: 0,
				EndTime:   10,
				TenantId:  "tenant_id",
			},
		},
	}
	fun := p.RegisterInitializeOp()
	httpHeader := transport.Header{}
	httpHeader.Set("org", "orgname")

	for _, comName := range []string{
		cpu,
		memory,
		diskIO,
		network,
	} {
		sdk := cptype.SDK{
			Ctx: transport.WithHeader(context.Background(), httpHeader),
			Comp: &cptype.Component{
				Name: comName,
			},
		}
		fun(&sdk)
	}
}
