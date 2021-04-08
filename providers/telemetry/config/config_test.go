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

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DefaultConfig(t *testing.T) {
	config := GlobalConfig()
	assert.Equal(t, config.QueryConfig.MonitorAddr, "localhost:7096")
}

func Test_InitConfig(t *testing.T) {
	Init(&Config{
		ReportConfig: &ReportConfig{
			Mode: STRICT_MODE,
			Collector: &CollectorConfig{
				Addr: "collector.default.svc.cluster.local:7076",
			},
		},
	})
	config := GlobalConfig()
	assert.Equal(t, config.ReportConfig.Collector.Addr, "collector.default.svc.cluster.local:7076")

	Init(&Config{
		ReportConfig: &ReportConfig{
			Mode:    PERFORMANCE_MODE,
			UdpHost: "127.0.0.1",
		},
	})

	config = GlobalConfig()
	assert.Equal(t, config.ReportConfig.Collector.Addr, "localhost:7076")
	assert.Equal(t, config.ReportConfig.UdpHost, "127.0.0.1")
}
