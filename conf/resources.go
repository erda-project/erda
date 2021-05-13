//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package conf

import _ "embed"

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
)
