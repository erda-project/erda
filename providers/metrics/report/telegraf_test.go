package report

import (
	"bou.ke/monkey"
	"github.com/erda-project/erda/providers/metrics/common"
	"gotest.tools/assert"
	"net"
	"os"
	"testing"
)

func Test_telegrafReporter_Send(t *testing.T) {
	telegraf := new(telegrafReporter)
	telegraf.conn, _ = net.Dial("tcp", os.Getenv("ADDR"))
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
	monkey.Patch(net.Conn.Write, func(conn net.Conn, b []byte) (n int, err error) {
		return 0, nil
	})
	err := telegraf.Send(metrics)
	assert.Equal(t, err, nil)
}
