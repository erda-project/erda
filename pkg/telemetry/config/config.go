package config

import (
	"os"
	"strconv"
	"sync"
)

type ReportMode string

var (
	PERFORMANCE_MODE ReportMode = "performance"
	STRICT_MODE      ReportMode = "strict"

	hostIpEnv            = "HOST_IP"
	collectorAddrEnv     = "COLLECTOR_ADDR"
	collectorUserNameEnv = "COLLECTOR_AUTH_USERNAME"
	collectorPasswordEnv = "COLLECTOR_AUTH_PASSWORD"
	collectorRetryEnv    = "TELEMETRY_REPORT_STRICT_RETRY"
	monitorAddrEnv       = "MONITOR_ADDR"
	reportModeEnv        = "TELEMETRY_REPORT_MODE"
	reportBufferSizeEnv  = "TELEMETRY_REPORT_BUFFER_SIZE"

	reportBufferSizeDefault = 200
	collectorRetryDefault   = 3
	collectorAddrDefault    = "localhost:7076"
	monitorAddrDefault      = "localhost:7096"
	hostIpDefault           = "localhost"
	hostPortDefault         = "7082"

	globalConfig = defaultConfig()
	globalMutex  sync.Mutex
)

type Config struct {
	ReportConfig *ReportConfig
	QueryConfig  *QueryConfig
}

type ReportConfig struct {
	Mode    ReportMode
	UdpHost string
	UdpPort string

	Collector *CollectorConfig

	BufferSize int
}

type CollectorConfig struct {
	Addr     string
	UserName string
	Password string
	Retry    int
}

type QueryConfig struct {
	MonitorAddr string
}

func GlobalConfig() *Config {
	return globalConfig
}

func Init(config *Config) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalConfig = defaultConfig()
	initOrMergeConfig(config)
}
func defaultConfig() *Config {
	config := &Config{
		ReportConfig: &ReportConfig{
			Mode:    PERFORMANCE_MODE,
			UdpHost: hostIpDefault,
			UdpPort: hostPortDefault,
			Collector: &CollectorConfig{
				Addr:  collectorAddrDefault,
				Retry: collectorRetryDefault,
			},
			BufferSize: reportBufferSizeDefault,
		},
		QueryConfig: &QueryConfig{
			MonitorAddr: monitorAddrDefault,
		},
	}
	monitorAddr := os.Getenv(monitorAddrEnv)
	if monitorAddr != "" {
		config.QueryConfig.MonitorAddr = monitorAddr
	}
	reportMode := os.Getenv(reportModeEnv)
	if reportMode != "" {
		config.ReportConfig.Mode = ReportMode(reportMode)
	}
	reportBufferSize := os.Getenv(reportBufferSizeEnv)
	if reportBufferSize != "" {
		if bufferSize, err := strconv.Atoi(reportBufferSize); err == nil {
			config.ReportConfig.BufferSize = bufferSize
		}
	}
	if config.ReportConfig.Mode == STRICT_MODE {
		collectorAddr := os.Getenv(collectorAddrEnv)
		if collectorAddr != "" {
			config.ReportConfig.Collector.Addr = collectorAddr
		}
		collectorUserName := os.Getenv(collectorUserNameEnv)
		if collectorUserName != "" {
			config.ReportConfig.Collector.UserName = collectorUserName
		}
		collectorPassword := os.Getenv(collectorPasswordEnv)
		if collectorPassword != "" {
			config.ReportConfig.Collector.Password = collectorPassword
		}
		collectorRetry := os.Getenv(collectorRetryEnv)
		if collectorRetry != "" {
			if retry, err := strconv.Atoi(collectorRetry); err == nil {
				config.ReportConfig.Collector.Retry = retry
			}
		}
	}
	if config.ReportConfig.Mode == PERFORMANCE_MODE {
		hostIp := os.Getenv(hostIpEnv)
		if hostIp != "" {
			config.ReportConfig.UdpHost = hostIp
		}
	}
	return config
}

func initOrMergeConfig(other *Config) {
	if other.QueryConfig != nil {
		if other.QueryConfig.MonitorAddr != "" {
			globalConfig.QueryConfig.MonitorAddr = other.QueryConfig.MonitorAddr
		}
	}
	if other.ReportConfig != nil {
		if other.ReportConfig.Mode != "" {
			globalConfig.ReportConfig.Mode = other.ReportConfig.Mode
		}
		if globalConfig.ReportConfig.Mode == STRICT_MODE {
			if other.ReportConfig.Collector != nil {
				if other.ReportConfig.Collector.Addr != "" {
					globalConfig.ReportConfig.Collector.Addr = other.ReportConfig.Collector.Addr
				}
				if other.ReportConfig.Collector.UserName != "" {
					globalConfig.ReportConfig.Collector.UserName = other.ReportConfig.Collector.UserName
				}
				if other.ReportConfig.Collector.Password != "" {
					globalConfig.ReportConfig.Collector.Password = other.ReportConfig.Collector.Password
				}
				if other.ReportConfig.Collector.Retry > 0 {
					globalConfig.ReportConfig.Collector.Retry = other.ReportConfig.Collector.Retry
				}
			}
		}
		if globalConfig.ReportConfig.Mode == PERFORMANCE_MODE {
			if other.ReportConfig.UdpHost != "" {
				globalConfig.ReportConfig.UdpHost = other.ReportConfig.UdpHost
			}
		}
		if other.ReportConfig.BufferSize > 0 {
			globalConfig.ReportConfig.BufferSize = other.ReportConfig.BufferSize
		}
	}
}
