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

package regex

//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
//
// 	"github.com/erda-project/erda/modules/monitor/core/metrics"
// )
//
// func testAndPrint(keys []*metrics.FieldDefine, metric, pattern, content string) {
// 	cfg, _ := json.Marshal(map[string]interface{}{
// 		"pattern": pattern,
// 		"keys":    keys,
// 	})
// 	p, err := New(metric, cfg)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	name, fields, err := p.Process(content)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println("name: ", name)
// 	for _, key := range keys {
// 		val := fields[key.Key]
// 		typ := reflect.TypeOf(val)
// 		fmt.Printf("%s (%s) = %v\n", key.Key, typ.Kind(), val)
// 	}
// }
//
// func ExampleProcessor() {
// 	keys := []*metrics.FieldDefine{
// 		&metrics.FieldDefine{
// 			Key:  "ip",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "time",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "method",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "url",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "request_time",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "request_length",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "status",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "length",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "ref_url",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "browser",
// 			Type: "string",
// 		},
// 	}
// 	testAndPrint(
// 		keys,
// 		"metric_name",
// 		"([\\d\\.]+) \\S+ \\S+ \\[(\\S+) \\S+\\] \"(\\w+) ([^\\\"]*)\" ([\\d\\.]+) (\\d+) (\\d+) (\\d+|-) \"([^\\\"]*)\" \"([^\\\"]*)\"",
// 		"10.200.0.101 - - [10/Aug/2017:14:57:51 +0800] \"POST /PutData?Category=YunOsAccountOpLog&AccessKeyId=abcdef&Date=Fri%2C%2028%20Jun%202013%2006%3A53%3A30%20GMT&Topic=raw&Signature=defg HTTP/1.1\" 0.024 18204 200 37 \"-\" \"aliyun-sdk-java\"",
// 	)
//
// 	// Output:
// 	// name:  metric_name
// 	// ip (string) = 10.200.0.101
// 	// time (string) = 10/Aug/2017:14:57:51
// 	// method (string) = POST
// 	// url (string) = /PutData?Category=YunOsAccountOpLog&AccessKeyId=abcdef&Date=Fri%2C%2028%20Jun%202013%2006%3A53%3A30%20GMT&Topic=raw&Signature=defg HTTP/1.1
// 	// request_time (string) = 0.024
// 	// request_length (float64) = 18204
// 	// status (string) = 200
// 	// length (float64) = 37
// 	// ref_url (string) = -
// 	// browser (string) = aliyun-sdk-java
// }
//
// func ExampleProcessor_nginx() {
// 	keys := []*metrics.FieldDefine{
// 		&metrics.FieldDefine{
// 			Key:  "userIp",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "scheme",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "host",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "requestLine",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "status",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "latency",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "requestSize",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "responseSize",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "referer",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "ua",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "upstreamAddr",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "upstreamStatus",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "upstreamLatency",
// 			Type: "number",
// 		},
// 	}
// 	testAndPrint(
// 		keys,
// 		"metric_name",
// 		`^[^\t]*\t([^\t,]*)[^\t]*\t[^\t]*\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t([^\t]*)\t[^\t]*\t[^\t]*\t[^\t]*\t([^\t]*)\t([^\t,]*)[^\t]*\t([^\t,]*)[^\t]*\t.*`,
// 		"[21/Jul/2020:15:39:24 +0800]\t127.0.0.1, 42.120.75.141\t10.118.183.0\thttp\tterminus-org.dev.terminus.io\tGET /api/deployments/actions/list-pending-approval?pageSize=15&type=BUILD&pageNo=1 HTTP/1.1\t200\t4.608\t1307\t2309\thttp://terminus-org.dev.terminus.io/workBench/approval/my-approve/pending?operator[]=1000019&pageNo=1&type=BUILD\tMozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36\t-\t-\t-\t10.96.239.163:9529\t200\t4.608\t20571\t1",
// 	)
//
// 	// Output:
// 	// name:  metric_name
// 	// userIp (string) = 127.0.0.1
// 	// scheme (string) = http
// 	// host (string) = terminus-org.dev.terminus.io
// 	// requestLine (string) = GET /api/deployments/actions/list-pending-approval?pageSize=15&type=BUILD&pageNo=1 HTTP/1.1
// 	// status (float64) = 200
// 	// latency (float64) = 4.608
// 	// requestSize (float64) = 1307
// 	// responseSize (float64) = 2309
// 	// referer (string) = http://terminus-org.dev.terminus.io/workBench/approval/my-approve/pending?operator[]=1000019&pageNo=1&type=BUILD
// 	// ua (string) = Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36
// 	// upstreamAddr (string) = 10.96.239.163:9529
// 	// upstreamStatus (float64) = 200
// 	// upstreamLatency (float64) = 4.608
// }
//
// func ExampleProcessor_tomcat() {
// 	keys := []*metrics.FieldDefine{
// 		&metrics.FieldDefine{
// 			Key:  "remoteIP",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "time",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "method",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "path",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "version",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "httpCode",
// 			Type: "number",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "group1",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "group2",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "ua",
// 			Type: "string",
// 		},
// 	}
// 	testAndPrint(
// 		keys,
// 		"metric_name",
// 		`^(?P<remoteIP>\S+) \S+ \S+ \[(?P<time>[\w:\/]+\s[+\-]\d{4})\] "(?P<method>\S+)\s?(?P<path>\S+)?\s?(?P<version>\S+)?" (?P<httpCode>\d{3}|-) (\d+|-)\s?"?([^"]*)"?\s?"?(?P<client>[^"]*)?"?$`,
// 		`11.11.11.11 - - [25/Jan/2000:14:00:01 +0100] "GET /1986.js HTTP/1.1" 200 932 "-" "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6"`,
// 	)
//
// 	// Output:
// 	// name:  metric_name
// 	// remoteIP (string) = 11.11.11.11
// 	// time (string) = 25/Jan/2000:14:00:01 +0100
// 	// method (string) = GET
// 	// path (string) = /1986.js
// 	// version (string) = HTTP/1.1
// 	// httpCode (float64) = 200
// 	// group1 (string) = 932
// 	// group2 (string) = -
// 	// ua (string) = Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6
// }
//
// func ExampleProcessor_t() {
// 	// tmall-fe
// 	keys := []*metrics.FieldDefine{
// 		&metrics.FieldDefine{
// 			Key:  "time",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "level",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "method",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "url",
// 			Type: "string",
// 		},
// 		&metrics.FieldDefine{
// 			Key:  "status",
// 			Type: "number",
// 		},
// 	}
// 	testAndPrint(
// 		keys,
// 		"metric_name",
// 		`^\[([^\]]+)\] (\S+) \[[^\]]+\] Proxy request \[([^:]+):([^\]]+)\] .* \[(\d+)\].+$`,
// 		"[Sun Jul 26 2020 12:41:01 GMT+0800 (China Standard Time)] WARNING [Middleware proxy] Proxy request [GET:/api/tick?count=5] get error status [404] ",
// 	)
//
// 	// Output:
// 	// name:  metric_name
// 	// time (string) = Sun Jul 26 2020 12:41:01 GMT+0800 (China Standard Time)
// 	// level (string) = WARNING
// 	// method (string) = GET
// 	// url (string) = /api/tick?count=5
// 	// status (float64) = 404
// }
