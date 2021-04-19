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

package slsimport

import (
	"bytes"
	"sort"
	"strconv"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	logs2 "github.com/erda-project/erda/modules/monitor/core/logs"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

// // RDS审计日志
// __topic__	日志主题，固定为rds_audit_log
// instance_id	RDS实例ID
// check_rows	扫描的行数
// db	数据库名
// fail	SQL 执行是否出错。0：成功，1：失败
// client_ip	访问RDS实例的客户端IP
// latency	延迟，单位为微秒
// origin_time	操作时间，单位为微秒
// return_rows	返回行数
// sql	执行的SQL语句
// thread_id	线程ID
// user	执行SQL的用户名
// update_rows	更新行数
//
func (c *Consumer) rdsProcess(shardID int, groups *sls.LogGroupList) string {
	product := "rds"
	buf := &bytes.Buffer{}
	filters := logFilterMap["rds"]

	for _, group := range groups.LogGroups {
		logs := group.Logs
		if len(filters) > 0 {
			for _, f := range filters {
				logs = f.FilterSLSLog(group.Logs)
			}
		}
		for _, log := range logs {
			timestamp := int64(log.GetTime()) * int64(time.Second)
			// 导入日志
			logTags := c.getTagsByContents(group, log.Contents, "")
			logTags["product"] = product
			// write time
			buf.WriteString(convertTimestampSecondToTimeString(int64(log.GetTime()), "") + " ")
			c.outputs.es.Write(&elasticsearch.Document{
				Index: c.getIndex(c.project, timestamp),
				Data: &logs2.Log{
					ID:        c.id,
					Source:    "sls",
					Stream:    "stdout",
					Offset:    (int64(shardID) << 32) + int64(log.GetTime()),
					Content:   getRDSContent(buf, log.Contents),
					Timestamp: timestamp,
					Tags:      logTags,
				},
			})

			tags := c.newMetricTags(group, product)
			// 导入指标
			m, err := getRDSMetrics(timestamp, tags, log.Contents)
			if err == nil {
				c.outputs.kafka.Write(m)
			}
		}
	}
	return ""
}

func getRDSContent(buf *bytes.Buffer, contents []*sls.LogContent) string {
	defer buf.Reset()
	sort.Sort(Contents(contents))
	for _, content := range contents {
		buf.WriteString(content.GetKey() + ":" + content.GetValue() + " ")
	}
	return buf.String()
}

func getRDSMetrics(timestamp int64, tags map[string]string, contents []*sls.LogContent) (m *metrics.Metric, err error) {
	fields := make(map[string]interface{})
	fields["count"] = 1
	for _, kv := range contents {
		switch kv.GetKey() {
		case "latency", "return_rows", "update_rows", "check_rows":
			if len(kv.GetValue()) > 0 {
				val, err := strconv.ParseInt(kv.GetValue(), 10, 64)
				if err != nil {
					return nil, err
				}
				fields[kv.GetKey()] = val
			}
		default:
			tags[kv.GetKey()] = kv.GetValue()
		}
	}
	return &metrics.Metric{
		Name:      "ali_rds_access",
		Timestamp: timestamp,
		Tags:      tags,
		Fields:    fields,
	}, nil
	return
}

func convertTimestampSecondToTimeString(t int64, layout string) string {
	tm := time.Unix(t, 0)
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return tm.Format(layout)
}
