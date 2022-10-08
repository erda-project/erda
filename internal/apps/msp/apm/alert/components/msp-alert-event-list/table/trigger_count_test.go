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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

type activeInfo struct {
	activeCount   int
	activeRequest []interface{}
}

type mockMetricService struct {
	request []*metricpb.QueryWithInfluxFormatRequest
	active  *activeInfo
}

func newMockMetricService() mockMetricService {
	return mockMetricService{
		active: &activeInfo{},
	}

}
func (m mockMetricService) QueryWithInfluxFormat(ctx context.Context, request *metricpb.QueryWithInfluxFormatRequest) (*metricpb.QueryWithInfluxFormatResponse, error) {
	if m.active == nil {
		m.active = &activeInfo{}
	}
	m.active.activeCount++
	m.active.activeRequest = append(m.active.activeRequest, request)
	return nil, nil
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

func TestFetchCount(t *testing.T) {
	step = 3

	tests := []struct {
		name      string
		mockEvent []*pb.AlertEventItem
		check     func(*testing.T, *activeInfo)
	}{
		{
			name:      "zero",
			mockEvent: []*pb.AlertEventItem{},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 0, info.activeCount)
			},
		},
		{
			name: "1",
			mockEvent: []*pb.AlertEventItem{
				{
					Id: "1",
				},
			},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 1, info.activeCount)
			},
		},
		{
			name: "3",
			mockEvent: []*pb.AlertEventItem{
				{
					Id: "1",
				},
				{
					Id: "2",
				},
				{
					Id: "3",
				},
			},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 1, info.activeCount)
			},
		},
		{
			name: "4",
			mockEvent: []*pb.AlertEventItem{
				{
					Id: "1",
				},
				{
					Id: "2",
				},
				{
					Id: "3",
				},
				{
					Id: "4",
				},
			},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 2, info.activeCount)
			},
		},
		{
			name: "5",
			mockEvent: []*pb.AlertEventItem{
				{
					Id: "1",
				},
				{
					Id: "2",
				},
				{
					Id: "3",
				},
				{
					Id: "4",
				},
				{
					Id: "5",
				},
			},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 2, info.activeCount)
			},
		},
		{
			name: "6",
			mockEvent: []*pb.AlertEventItem{
				{
					Id: "1",
				},
				{
					Id: "2",
				},
				{
					Id: "3",
				},
				{
					Id: "4",
				},
				{
					Id: "5",
				},
				{
					Id: "6",
				},
			},
			check: func(t *testing.T, info *activeInfo) {
				require.Equal(t, 2, info.activeCount)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMetric := newMockMetricService()
			triggerCount := TriggerCount{
				events: test.mockEvent,
				metric: mockMetric,
			}
			err := triggerCount.Fetch()
			require.NoError(t, err)
			test.check(t, mockMetric.active)

		})
	}
}
