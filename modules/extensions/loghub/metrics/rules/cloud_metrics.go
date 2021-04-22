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

package rules

// import "terminus.io/dice/monitor/modules/monitor/metrics"

// var (
// 	cloudProducts = []string{"ali_waf_access", "ali_api_gateway_access"}
// 	cloudMetrics  = map[string]*metrics.MetricMeta{
// 		"ali_waf_access": &metrics.MetricMeta{
// 			Name: metrics.NameDefine{
// 				Key:  "ali_waf_access",
// 				Name: "WAF Access",
// 			},
// 			Tags: map[string]*metrics.TagDefine{
// 				"ua_os": &metrics.TagDefine{
// 					Key:  "ua_os",
// 					Name: "UA OS",
// 				},
// 				"remote_addr": &metrics.TagDefine{
// 					Key:  "remote_addr",
// 					Name: "Remote Addr",
// 				},
// 				"request_method": &metrics.TagDefine{
// 					Key:  "request_method",
// 					Name: "Request Method",
// 				},
// 				"sls_log_store": &metrics.TagDefine{
// 					Key:  "sls_log_store",
// 					Name: "SLS Log Store",
// 				},
// 				"ua_browser_type": &metrics.TagDefine{
// 					Key:  "ua_browser_type",
// 					Name: "UA Browser Type",
// 				},
// 				"org_name": &metrics.TagDefine{
// 					Key:  "org_name",
// 					Name: "Org Name",
// 				},
// 				"server_protocol": &metrics.TagDefine{
// 					Key:  "server_protocol",
// 					Name: "Server Protocol",
// 				},
// 				"status": &metrics.TagDefine{
// 					Key:  "status",
// 					Name: "Status",
// 				},
// 				"host": &metrics.TagDefine{
// 					Key:  "host",
// 					Name: "Host",
// 				},
// 				"http_referer": &metrics.TagDefine{
// 					Key:  "http_referer",
// 					Name: "HTTP Referer",
// 				},
// 				"https": &metrics.TagDefine{
// 					Key:  "https",
// 					Name: "Https",
// 				},
// 				"upstream_addr": &metrics.TagDefine{
// 					Key:  "upstream_addr",
// 					Name: "Upstream Addr",
// 				},
// 				"http_error": &metrics.TagDefine{
// 					Key:  "http_error",
// 					Name: "HTTP Error",
// 				},
// 				"http_x_forwarded_for": &metrics.TagDefine{
// 					Key:  "http_x_forwarded_for",
// 					Name: "HTTP X Forwarded For",
// 				},
// 				"product": &metrics.TagDefine{
// 					Key:  "product",
// 					Name: "Product",
// 				},
// 				"real_client_ip": &metrics.TagDefine{
// 					Key:  "real_client_ip",
// 					Name: "Real Client IP",
// 				},
// 				"sls_topic": &metrics.TagDefine{
// 					Key:  "sls_topic",
// 					Name: "SLS Topic",
// 				},
// 				"ua_browser_family": &metrics.TagDefine{
// 					Key:  "ua_browser_family",
// 					Name: "UA Browser Family",
// 				},
// 				"upstream_response_time": &metrics.TagDefine{
// 					Key:  "upstream_response_time",
// 					Name: "Upstream Response Time",
// 				},
// 				"block_action": &metrics.TagDefine{
// 					Key:  "block_action",
// 					Name: "Block Action",
// 				},
// 				"dice_org_name": &metrics.TagDefine{
// 					Key:  "dice_org_name",
// 					Name: "Dice Org Name",
// 				},
// 				"request_path": &metrics.TagDefine{
// 					Key:  "request_path",
// 					Name: "Request Path",
// 				},
// 				"request_traceid": &metrics.TagDefine{
// 					Key:  "request_traceid",
// 					Name: "Request Traceid",
// 				},
// 				"dice_org_id": &metrics.TagDefine{
// 					Key:  "dice_org_id",
// 					Name: "Dice Org ID",
// 				},
// 				"sls_source": &metrics.TagDefine{
// 					Key:  "sls_source",
// 					Name: "SLS Source",
// 				},
// 				"status_type": &metrics.TagDefine{
// 					Key:  "status_type",
// 					Name: "Status Type",
// 				},
// 				"wxbb_invalid_wua": &metrics.TagDefine{
// 					Key:  "wxbb_invalid_wua",
// 					Name: "Wxbb Invalid Wua",
// 				},
// 				"http_user_agent": &metrics.TagDefine{
// 					Key:  "http_user_agent",
// 					Name: "HTTP User Agent",
// 				},
// 				"sls_category": &metrics.TagDefine{
// 					Key:  "sls_category",
// 					Name: "SLS Category",
// 				},
// 				"sls_project": &metrics.TagDefine{
// 					Key:  "sls_project",
// 					Name: "SLS Project",
// 				},
// 				"ua_browser": &metrics.TagDefine{
// 					Key:  "ua_browser",
// 					Name: "UA Browser",
// 				},
// 				"matched_host": &metrics.TagDefine{
// 					Key:  "matched_host",
// 					Name: "Matched Host",
// 				},
// 				"origin": &metrics.TagDefine{
// 					Key:  "origin",
// 					Name: "Origin",
// 				},
// 				"user_id": &metrics.TagDefine{
// 					Key:  "user_id",
// 					Name: "User ID",
// 				},
// 				"ua_os_family": &metrics.TagDefine{
// 					Key:  "ua_os_family",
// 					Name: "UA OS Family",
// 				},
// 				"wxbb_rule_id": &metrics.TagDefine{
// 					Key:  "wxbb_rule_id",
// 					Name: "Wxbb Rule ID",
// 				},
// 				"wxbb_test": &metrics.TagDefine{
// 					Key:  "wxbb_test",
// 					Name: "Wxbb Test",
// 				},
// 				"content_type": &metrics.TagDefine{
// 					Key:  "content_type",
// 					Name: "Content Type",
// 				},
// 				"region": &metrics.TagDefine{
// 					Key:  "region",
// 					Name: "Region",
// 				},
// 				"time": &metrics.TagDefine{
// 					Key:  "time",
// 					Name: "Time",
// 				},
// 				"ua_device_type": &metrics.TagDefine{
// 					Key:  "ua_device_type",
// 					Name: "UA Device Type",
// 				},
// 			},
// 			Fields: map[string]*metrics.FieldDefine{
// 				"body_bytes_sent": &metrics.FieldDefine{
// 					Key:  "body_bytes_sent",
// 					Type: "number",
// 					Name: "Body Bytes Sent",
// 					Unit: "",
// 				},
// 				"remote_port": &metrics.FieldDefine{
// 					Key:  "remote_port",
// 					Type: "number",
// 					Name: "Remote Port",
// 					Unit: "",
// 				},
// 				"request_length": &metrics.FieldDefine{
// 					Key:  "request_length",
// 					Type: "number",
// 					Name: "Request Length",
// 					Unit: "",
// 				},
// 				"request_time_msec": &metrics.FieldDefine{
// 					Key:  "request_time_msec",
// 					Type: "number",
// 					Name: "Request Time Msec",
// 					Unit: "",
// 				},
// 				"status": &metrics.FieldDefine{
// 					Key:  "status",
// 					Type: "number",
// 					Name: "Status",
// 					Unit: "",
// 				},
// 				"upstream_status": &metrics.FieldDefine{
// 					Key:  "upstream_status",
// 					Type: "number",
// 					Name: "Upstream Status",
// 					Unit: "",
// 				},
// 				"count": &metrics.FieldDefine{
// 					Key:  "count",
// 					Type: "number",
// 					Name: "Count",
// 					Unit: "",
// 				},
// 			},
// 		},
// 		"ali_api_gateway_access": &metrics.MetricMeta{
// 			Name: metrics.NameDefine{
// 				Key:  "ali_api_gateway_access",
// 				Name: "API Gateway Access",
// 			},
// 			Tags: map[string]*metrics.TagDefine{
// 				"path": &metrics.TagDefine{
// 					Key:  "path",
// 					Name: "Path",
// 				},
// 				"region": &metrics.TagDefine{
// 					Key:  "region",
// 					Name: "Region",
// 				},
// 				"appName": &metrics.TagDefine{
// 					Key:  "appName",
// 					Name: "App Name",
// 				},
// 				"consumerAppKey": &metrics.TagDefine{
// 					Key:  "consumerAppKey",
// 					Name: "Consumer App Key",
// 				},
// 				"dice_org_name": &metrics.TagDefine{
// 					Key:  "dice_org_name",
// 					Name: "Dice Org Name",
// 				},
// 				"httpMethod": &metrics.TagDefine{
// 					Key:  "httpMethod",
// 					Name: "HTTP Method",
// 				},
// 				"providerAliUid": &metrics.TagDefine{
// 					Key:  "providerAliUid",
// 					Name: "Provider Ali UID",
// 				},
// 				"requestHandleTime": &metrics.TagDefine{
// 					Key:  "requestHandleTime",
// 					Name: "Request Handle Time",
// 				},
// 				"apiStageName": &metrics.TagDefine{
// 					Key:  "apiStageName",
// 					Name: "API Stage Name",
// 				},
// 				"sls_project": &metrics.TagDefine{
// 					Key:  "sls_project",
// 					Name: "SLS Project",
// 				},
// 				"responseHeaders": &metrics.TagDefine{
// 					Key:  "responseHeaders",
// 					Name: "Response Headers",
// 				},
// 				"sls_category": &metrics.TagDefine{
// 					Key:  "sls_category",
// 					Name: "SLS Category",
// 				},
// 				"exception": &metrics.TagDefine{
// 					Key:  "exception",
// 					Name: "Exception",
// 				},
// 				"initialRequestId": &metrics.TagDefine{
// 					Key:  "initialRequestId",
// 					Name: "Initial Request ID",
// 				},
// 				"sls_topic": &metrics.TagDefine{
// 					Key:  "sls_topic",
// 					Name: "SLS Topic",
// 				},
// 				"statusCode": &metrics.TagDefine{
// 					Key:  "statusCode",
// 					Name: "Status Code",
// 				},
// 				"statusCodeType": &metrics.TagDefine{
// 					Key:  "statusCodeType",
// 					Name: "Status Code Type",
// 				},
// 				"apiGroupUid": &metrics.TagDefine{
// 					Key:  "apiGroupUid",
// 					Name: "API Group UID",
// 				},
// 				"clientNonce": &metrics.TagDefine{
// 					Key:  "clientNonce",
// 					Name: "Client Nonce",
// 				},
// 				"product": &metrics.TagDefine{
// 					Key:  "product",
// 					Name: "Product",
// 				},
// 				"requestProtocol": &metrics.TagDefine{
// 					Key:  "requestProtocol",
// 					Name: "Request Protocol",
// 				},
// 				"sls_source": &metrics.TagDefine{
// 					Key:  "sls_source",
// 					Name: "SLS Source",
// 				},
// 				"http_error": &metrics.TagDefine{
// 					Key:  "http_error",
// 					Name: "HTTP Error",
// 				},
// 				"domain": &metrics.TagDefine{
// 					Key:  "domain",
// 					Name: "Domain",
// 				},
// 				"errorMessage": &metrics.TagDefine{
// 					Key:  "errorMessage",
// 					Name: "Error Message",
// 				},
// 				"apiName": &metrics.TagDefine{
// 					Key:  "apiName",
// 					Name: "API Name",
// 				},
// 				"appId": &metrics.TagDefine{
// 					Key:  "appId",
// 					Name: "App ID",
// 				},
// 				"requestHeaders": &metrics.TagDefine{
// 					Key:  "requestHeaders",
// 					Name: "Request Headers",
// 				},
// 				"apiGroupName": &metrics.TagDefine{
// 					Key:  "apiGroupName",
// 					Name: "API Group Name",
// 				},
// 				"origin": &metrics.TagDefine{
// 					Key:  "origin",
// 					Name: "Origin",
// 				},
// 				"requestQueryString": &metrics.TagDefine{
// 					Key:  "requestQueryString",
// 					Name: "Request Query String",
// 				},
// 				"sls_log_store": &metrics.TagDefine{
// 					Key:  "sls_log_store",
// 					Name: "SLS Log Store",
// 				},
// 				"apiStageUid": &metrics.TagDefine{
// 					Key:  "apiStageUid",
// 					Name: "API Stage UID",
// 				},
// 				"dice_org_id": &metrics.TagDefine{
// 					Key:  "dice_org_id",
// 					Name: "Dice Org ID",
// 				},
// 				"customTraceId": &metrics.TagDefine{
// 					Key:  "customTraceId",
// 					Name: "Custom Trace ID",
// 				},
// 				"errorCode": &metrics.TagDefine{
// 					Key:  "errorCode",
// 					Name: "Error Code",
// 				},
// 				"requestId": &metrics.TagDefine{
// 					Key:  "requestId",
// 					Name: "Request ID",
// 				},
// 				"apiUid": &metrics.TagDefine{
// 					Key:  "apiUid",
// 					Name: "API UID",
// 				},
// 				"clientIp": &metrics.TagDefine{
// 					Key:  "clientIp",
// 					Name: "Client IP",
// 				},
// 				"requestBody": &metrics.TagDefine{
// 					Key:  "requestBody",
// 					Name: "Request Body",
// 				},
// 				"responseBody": &metrics.TagDefine{
// 					Key:  "responseBody",
// 					Name: "Response Body",
// 				},
// 				"instanceId": &metrics.TagDefine{
// 					Key:  "instanceId",
// 					Name: "Instance ID",
// 				},
// 				"org_name": &metrics.TagDefine{
// 					Key:  "org_name",
// 					Name: "Org Name",
// 				},
// 			},
// 			Fields: map[string]*metrics.FieldDefine{
// 				"requestSize": &metrics.FieldDefine{
// 					Key:  "requestSize",
// 					Type: "number",
// 					Name: "Request Size",
// 					Unit: "",
// 				},
// 				"responseSize": &metrics.FieldDefine{
// 					Key:  "responseSize",
// 					Type: "number",
// 					Name: "Response Size",
// 					Unit: "",
// 				},
// 				"serviceLatency": &metrics.FieldDefine{
// 					Key:  "serviceLatency",
// 					Type: "number",
// 					Name: "Service Latency",
// 					Unit: "",
// 				},
// 				"statusCode": &metrics.FieldDefine{
// 					Key:  "statusCode",
// 					Type: "number",
// 					Name: "Status Code",
// 					Unit: "",
// 				},
// 				"totalLatency": &metrics.FieldDefine{
// 					Key:  "totalLatency",
// 					Type: "number",
// 					Name: "Total Latency",
// 					Unit: "",
// 				},
// 				"count": &metrics.FieldDefine{
// 					Key:  "count",
// 					Type: "number",
// 					Name: "Count",
// 					Unit: "",
// 				},
// 			},
// 		},
// 	}
// )

// func (a *Adapt) listCloudMetricNames() []*metrics.NameDefine {
// 	var list []*metrics.NameDefine
// 	for _, p := range cloudProducts {
// 		list = append(list, &cloudMetrics[p].Name)
// 	}
// 	return list
// }

// func (a *Adapt) getCloudMetricDefine(metric string) *metrics.MetricMeta {
// 	return cloudMetrics[metric]
// }

// func (a *Adapt) listCloudMetricDefine() []*metrics.MetricMeta {
// 	var list []*metrics.MetricMeta
// 	for _, p := range cloudProducts {
// 		list = append(list, cloudMetrics[p])
// 	}
// 	return list
// }
