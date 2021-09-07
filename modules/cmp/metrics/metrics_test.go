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

package metrics

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/mock"
)

func TestMetricsSearch(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	server := mock.NewMockMetricServiceServer(ctl)
	mockServer := Metric{Metricq: server}
	var err error
	ctx := context.Background()

	//resp :=&pb.QueryWithInfluxFormatResponse{}

	req1 := &pb.QueryWithInfluxFormatRequest{
		Start:     "before_1h",
		End:       "now",
		Statement: NodeCpuUsageSelectStatement,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue("local"),
			"hostname":     structpb.NewStringValue("terminus-dev"),
		}}
	req2 := &pb.QueryWithInfluxFormatRequest{
		Start:     "before_1h",
		End:       "now",
		Statement: NodeMemoryUsageSelectStatement,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue("local"),
			"hostname":     structpb.NewStringValue("terminus-dev"),
		}}
	req3 := &pb.QueryWithInfluxFormatRequest{
		Start:     "before_1h",
		End:       "now",
		Statement: NodeCpuUsageSelectStatement,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue("local"),
			"hostname":     structpb.NewStringValue("terminus-dev"),
		}}
	gomock.InOrder(
		server.EXPECT().QueryWithInfluxFormat(ctx, req1),
		server.EXPECT().QueryWithInfluxFormat(ctx, req2),
		server.EXPECT().QueryWithInfluxFormat(ctx, req3),
	)

	//server.QueryWithInfluxFormat(ctx,&req1)
	_, err = mockServer.DoQuery(ctx, cache.GenerateKey([]string{req1.Params["hostname"].GetStringValue(), "local", "cpu"}), req1)
	if err != nil {
		return
	}

	_, err = mockServer.DoQuery(ctx, cache.GenerateKey([]string{req2.Params["hostname"].GetStringValue(), "local", "mem"}), req2)
	if err != nil {
		return
	}
	key := cache.GenerateKey([]string{req3.Params["hostname"].GetStringValue(), "local", "cpu"})
	cache.FreeCache.Remove(key)
	_, err = mockServer.DoQuery(ctx, key, req3)
	if err != nil {
		return
	}
}
