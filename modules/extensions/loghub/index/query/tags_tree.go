package query

import (
	"github.com/erda-project/erda-infra/providers/i18n"
)

func getScopeTagsTree(scope string) map[string]interface{} {
	if scope == "micro_service" {
		return map[string]interface{}{
			"key":       "dice_application_id",
			"dimension": "app",
		}
	}
	// org
	return map[string]interface{}{
		"key":       "dice_project_id",
		"dimension": "project",
		"dynamic_children": map[string]interface{}{
			"key":       "dice_application_id",
			"dimension": "app",
		},
	}
}

// GetTagsTree .
func (p *provider) GetTagsTree(scope string, lang i18n.LanguageCodes) []map[string]interface{} {
	return []map[string]interface{}{
		map[string]interface{}{
			"tag": map[string]interface{}{
				"name":  p.t.Text(lang, "Dice"),
				"key":   "origin",
				"value": "dice",
			},
			"dynamic_children": getScopeTagsTree(scope),
		},
		map[string]interface{}{
			"tag": map[string]interface{}{
				"name":  p.t.Text(lang, "Aliyun"),
				"key":   "origin",
				"value": "sls",
			},
			"children": []map[string]interface{}{
				map[string]interface{}{
					"tag": map[string]interface{}{
						"name":  p.t.Text(lang, "WAF"),
						"key":   "product",
						"value": "waf",
					},
					"children": []map[string]interface{}{
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Matched Host",
								"key":  "matched_host",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Real Client IP",
								"key":  "real_client_ip",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Region",
								"key":  "region",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Browser",
								"key":  "ua_browser",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "User ID",
								"key":  "user_id",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Block Action",
								"key":  "block_action",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Content Type",
								"key":  "content_type",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Os",
								"key":  "ua_os",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Upstream Addr",
								"key":  "upstream_addr",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Wxbb Rule ID",
								"key":  "wxbb_rule_id",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "HTTP Referer",
								"key":  "http_referer",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Traceid",
								"key":  "request_traceid",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Category",
								"key":  "sls_category",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "HTTP User Agent",
								"key":  "http_user_agent",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "HTTP X Forwarded For",
								"key":  "http_x_forwarded_for",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Browser Version",
								"key":  "ua_browser_version",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Device Type",
								"key":  "ua_device_type",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Upstream Response Time",
								"key":  "upstream_response_time",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Host",
								"key":  "host",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Remote Port",
								"key":  "remote_port",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Time",
								"key":  "time",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Browser Type",
								"key":  "ua_browser_type",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Wxbb Test",
								"key":  "wxbb_test",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Source",
								"key":  "sls_source",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Status",
								"key":  "status",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Browser Family",
								"key":  "ua_browser_family",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Wxbb Invalid Wua",
								"key":  "wxbb_invalid_wua",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Body Bytes Sent",
								"key":  "body_bytes_sent",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Path",
								"key":  "request_path",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Time Msec",
								"key":  "request_time_msec",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Server Protocol",
								"key":  "server_protocol",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Project",
								"key":  "sls_project",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Topic",
								"key":  "sls_topic",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Ua Os Family",
								"key":  "ua_os_family",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "HTTP Cookie",
								"key":  "http_cookie",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Https",
								"key":  "https",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Remote Addr",
								"key":  "remote_addr",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Length",
								"key":  "request_length",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Method",
								"key":  "request_method",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Log Store",
								"key":  "sls_log_store",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Upstream Status",
								"key":  "upstream_status",
							},
						},
					},
				},
				map[string]interface{}{
					"tag": map[string]interface{}{
						"name":  p.t.Text(lang, "APIGateway"),
						"key":   "product",
						"value": "apigateway",
					},
					"children": []map[string]interface{}{
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Error Message",
								"key":  "errorMessage",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Exception",
								"key":  "exception",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "HTTP Method",
								"key":  "httpMethod",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Headers",
								"key":  "requestHeaders",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Response Body",
								"key":  "responseBody",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API Group Name",
								"key":  "apiGroupName",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API Stage Name",
								"key":  "apiStageName",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Topic",
								"key":  "sls_topic",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Body",
								"key":  "requestBody",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Client Nonce",
								"key":  "clientNonce",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Consumer App Key",
								"key":  "consumerAppKey",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Provider Ali UID",
								"key":  "providerAliUid",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Instance ID",
								"key":  "instanceId",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Handle Time",
								"key":  "requestHandleTime",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Size",
								"key":  "requestSize",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Custom Trace ID",
								"key":  "customTraceId",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Initial Request ID",
								"key":  "initialRequestId",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Protocol",
								"key":  "requestProtocol",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request Query String",
								"key":  "requestQueryString",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Response Size",
								"key":  "responseSize",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Source",
								"key":  "sls_source",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Status Code",
								"key":  "statusCode",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API Group UID",
								"key":  "apiGroupUid",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "App ID",
								"key":  "appId",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Path",
								"key":  "path",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Total Latency",
								"key":  "totalLatency",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Response Headers",
								"key":  "responseHeaders",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Log Store",
								"key":  "sls_log_store",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Project",
								"key":  "sls_project",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API Name",
								"key":  "apiName",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API Stage UID",
								"key":  "apiStageUid",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Error Code",
								"key":  "errorCode",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Request ID",
								"key":  "requestId",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Service Latency",
								"key":  "serviceLatency",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "App Name",
								"key":  "appName",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Region",
								"key":  "region",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Domain",
								"key":  "domain",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "SLS Category",
								"key":  "sls_category",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "API UID",
								"key":  "apiUid",
							},
						},
						map[string]interface{}{
							"tag": map[string]interface{}{
								"name": "Client IP",
								"key":  "clientIp",
							},
						},
					},
				},
				// map[string]interface{}{
				// 	"tag": map[string]interface{}{
				// 		"name":  p.t.Text(lang, "NoSQL"),
				// 		"key":   "product",
				// 		"value": "nosql",
				// 	},
				// 	"children": []map[string]interface{}{
				// 		map[string]interface{}{
				// 			"tag": map[string]interface{}{
				// 				"name": p.t.Text(lang, "cmd"),
				// 				"key":  "cmd",
				// 			},
				// 		},
				// 	},
				// },
			},
		},
	}
}
