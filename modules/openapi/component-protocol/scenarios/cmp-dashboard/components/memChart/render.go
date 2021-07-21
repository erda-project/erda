package memChart

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

var (
	metricsServer = servicehub.New().Service("metrics-query").(pb.MetricServiceServer)
	defaultDuration = 24*time.Hour
)

// GenComponentState 获取state
func (chart *MemChart) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state common.State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	chart.State = state
	return nil
}
func (chart *MemChart) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	chart.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	var (
		resp *pb.QueryWithInfluxFormatResponse
		err  error
	)
	if err = chart.GenComponentState(c); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		chart.State.Start = time.Now().Truncate(defaultDuration)
		chart.State.End = time.Now()
	case apistructs.CMPDashboardChart:
		break
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	req := &pb.QueryWithInfluxFormatRequest{
		Start:   chart.State.Start.String(),
		End:     time.Now().String(),
		Filters: nil,
		Options: nil,
		Statement: `SELECT cpu_usage_active, timestamp FROM status_page 
	WHERE cluster_name::tag=$cluster_name && hostname::tag=$hostname 
	ORDER BY TIMESTAMP DESC`,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue(chart.State.ClusterName),
			"hostname":    structpb.NewStringValue(chart.State.Name),
		},
	}
	if resp, err = metricsServer.QueryWithInfluxFormat(context.Background(), req); err != nil {
		return err
	}
	var items []common.ChartDataItem
	for _, res := range resp.Results {
		for _, serie := range res.Series {
			for _,row:=range serie.Rows{
				v := row.Values[0].GetNumberValue()
				t := row.Values[1].GetNumberValue()
				items = append(items, common.ChartDataItem{
					Value: v,
					Time:  int64(t),
				})
			}
		}
	}
	chart.Data = items
	return nil
}
