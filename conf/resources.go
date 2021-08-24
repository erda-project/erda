// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import "embed"

var (
	//go:embed openapi-ng/openapi-ng.yaml
	OpenAPINGDefaultConfig  string
	OpenAPINGConfigFilePath = "conf/openapi-ng/openapi-ng.yaml"

	//go:embed openapi/openapi.yaml
	OpenAPIDefaultConfig  string
	OpenAPIConfigFilePath = "conf/openapi/openapi.yaml"
)

// monitor
var (
	//go:embed monitor/collector/collector.yaml
	MonitorCollectorDefaultConfig  string
	MonitorCollectorConfigFilePath string = "conf/monitor/collector/collector.yaml"

	//go:embed monitor/monitor/monitor.yaml
	MonitorDefaultConfig  string
	MonitorConfigFilePath string = "conf/monitor/monitor/monitor.yaml"

	//go:embed monitor/streaming/streaming.yaml
	MonitorStreamingDefaultConfig  string
	MonitorStreamingConfigFilePath string = "conf/monitor/streaming/streaming.yaml"

	// extensions
	//go:embed monitor/extensions/report-engine.yaml
	MonitorReportEngineDefaultConfig string
	MonitorReportEngineFilePath      string = "conf/monitor/extensions/report-engine.yaml"
)

// msp
var (
	//go:embed msp/msp.yaml
	MSPDefaultConfig  string
	MSPConfigFilePath string = "conf/msp/msp.yaml"

	//go:embed msp/init_sqls
	MSPAddonInitSqls   embed.FS
	MSPAddonFsRootPath = "msp/init_sqls"
)
