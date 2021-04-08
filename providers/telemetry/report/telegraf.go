// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package report

import (
	"encoding/json"
	"net"

	"github.com/erda-project/erda/providers/telemetry/common"
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

func newTelegrafReporter(hostIp string, port string) (*telegrafReporter, error) {
	conn, err := net.Dial("udp", hostIp+":"+port)
	if err != nil {
		return nil, err
	}
	return &telegrafReporter{conn: conn}, nil
}
