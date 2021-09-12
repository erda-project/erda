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

package common

const Layout = "2006-01-02 15:04:05"

type Void struct{}

const (
	// CallAnalysisHttpClient http-client
	CallAnalysisHttpClient = "span_call_analysis_http_client"
	// CallAnalysisHttpServer http-server
	CallAnalysisHttpServer = "span_call_analysis_http_server"
	// CallAnalysisRpcClient rpc-client
	CallAnalysisRpcClient = "span_call_analysis_rpc_client"
	// CallAnalysisRpcServer rpc-server
	CallAnalysisRpcServer = "span_call_analysis_rpc_server"
	// CallAnalysisCacheClient cache client
	CallAnalysisCacheClient = "span_call_analysis_cache_client"
	// CallAnalysisMqProducer mq producer
	CallAnalysisMqProducer = "span_call_analysis_mq_producer"
	// CallAnalysisMqConsumer mq consumer
	CallAnalysisMqConsumer = "span_call_analysis_mq_consumer"
	// CallAnalysisInvokeLocal invoke local
	CallAnalysisInvokeLocal = "span_call_analysis_invoke_local"
)
