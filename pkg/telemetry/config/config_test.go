package config

import (
	"gotest.tools/assert"
	"os"
	"testing"
)

func Test_DefaultConfig(t *testing.T) {
	config := GlobalConfig()
	assert.Equal(t, config.QueryConfig.MonitorAddr, "localhost:7096")
}

func Test_EnvConfig(t *testing.T) {
	_ = os.Setenv("MONITOR_ADDR", "monitor.default.svc.cluster.local:7086")
	config := GlobalConfig()
	assert.Equal(t, config.QueryConfig.MonitorAddr, "monitor.default.svc.cluster.local:7086")
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
