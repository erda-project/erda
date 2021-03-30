package report

import (
	"encoding/json"
	"net"

	"github.com/erda-project/erda/providers/metrics/common"
)

var DEFAULT_BUCKET = 10

type telegrafReporter struct {
	conn net.Conn
}

func (r *telegrafReporter) Send(metrics []*common.Metric) (err error) {
	length := len(metrics)
	if length == 0 {
		return
	}
	if length <= DEFAULT_BUCKET {
		return r.send(metrics)
	}
	idx := 0
	for {
		bucket := DEFAULT_BUCKET
		if length-idx < DEFAULT_BUCKET {
			bucket = length - idx
		}
		end := idx + bucket
		bucketMetrics := metrics[idx:end]
		err = r.send(bucketMetrics)
		idx = end
		if idx >= length {
			break
		}
	}
	return err
}

func (r *telegrafReporter) send(metrics []*common.Metric) (err error) {
	if data, err := json.Marshal(metrics); err == nil {
		if r.conn != nil {
			_, err = r.conn.Write(data)
		}
	}
	return err
}
