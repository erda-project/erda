package esinfluxql

import (
	"strings"
	"testing"

	"github.com/influxdata/influxql"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func TestElasticsearchParse(t *testing.T) {
	p := Parser{
		ql: influxql.NewParser(strings.NewReader("SELECT service_id::tag,type::tag,sum(count::field) FROM error_alert WHERE terminus_key::tag=$terminus_key AND service_id::tag=$service_id GROUP BY type::tag ORDER BY sum(count::field) DESC LIMIT 5")),
		ctx: &Context{
			start:            0,
			end:              0,
			originalTimeUnit: tsql.Nanosecond,
			targetTimeUnit:   tsql.UnsetTimeUnit,
			timeKey:          model.TimestampKey,
			maxTimePoints:    512,
		},
	}
	p.SetParams(map[string]interface{}{"terminus_key": "123", "service_id": "456"})
	//got, err := p.ParseQuery()
	//if err != nil {
	//	t.Error(err)
	//}
	//_ = got
}

func TestMetrics(t *testing.T) {
	tests := []struct {
		name   string
		stm    string
		want   []string
		params map[string]interface{}
		err    error
	}{
		{
			name: "single metric, normal stmt",
			stm:  "select * from application_http",
			want: []string{"application_http"},
		},
		{
			name: "two metric, normal stmt",
			stm:  "select * from application_http,application_rpc",
			want: []string{"application_http", "application_rpc"},
		},
		{
			name:   "two metric, real stmt",
			stm:    "SELECT sum(if(eq(error::tag, 'true'),elapsed_count::field,0))/sum(elapsed_count::field) FROM application_http,application_rpc WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) AND (target_service_id::tag=$service_id OR target_service_id::tag=$service_id)  GROUP BY time()",
			params: map[string]interface{}{"terminus_key": "123", "service_id": "456"},
			want:   []string{"application_http", "application_rpc"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(0, 0, test.stm, false)
			if test.params != nil {
				p.SetParams(test.params)
			}
			err := p.Build()
			require.NoError(t, err)
			got, err := p.Metrics()
			if err != test.err {
				t.Errorf("got error %v, want %v", err, test.err)
			}
			require.ElementsMatchf(t, got, test.want, "got %v, want %v", got, test.want)
		})
	}
}
