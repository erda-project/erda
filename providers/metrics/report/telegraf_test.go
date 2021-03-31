package report

import (
	"bou.ke/monkey"
	"github.com/erda-project/erda/providers/metrics/common"
	"gotest.tools/assert"
	"net"
	"testing"
)

func Test_telegrafReporter_Send(t *testing.T) {
	r := new(telegrafReporter)
	r.conn, _ = net.Dial("tcp", "monitor.default.svc.cluster.local:7096")
	writeResult := monkey.Patch(r.conn.Write, func(b []byte) (n int, err error) {
		return 0, nil
	})
	defer writeResult.Unpatch()
	metrics := []*common.Metric{
		{
			Name:      "_metric_meta",
			Timestamp: 1614583470000,
			Tags: map[string]string{
				"cluster_name": "terminus-dev",
				"meta":         "true",
				"metric_name":  "application_db",
			},
			Fields: map[string]interface{}{
				"fields": []string{"value:number"},
				"tags":   []string{"is_edge", "org_id"},
			},
		},
	}
	err := r.Send(metrics)
	assert.Equal(t, nil, err)
}
