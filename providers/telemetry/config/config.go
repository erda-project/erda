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
	"os"
	"strconv"
	"sync"
)

var (
	/*
		高性能模式
		在此选项下，使用UDP向 Telegraf DaemonSet 进行数据上报，但不保证事务和数据的不丢失。
	*/
	PERFORMANCE_MODE ReportMode = "performance"

	/*
		严格模式
		在此选项下，使用HTTP向 Collector 进行数据上报，并提供上报失败的重试机制，以保证数据不会丢失。
	*/
	STRICT_MODE ReportMode = "strict"

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

type ReportMode string

type ReportConfig struct {
	Mode    ReportMode
	UdpHost string
	UdpPort string

	Collector *CollectorConfig

	BufferSize int
}

type CollectorConfig struct {
	Addr     string `file:"addr" env:"COLLECTOR_ADDR" default:"localhost:7076"`
	UserName string `file:"username" env:"COLLECTOR_AUTH_USERNAME"`
	Password string `file:"password" env:"COLLECTOR_AUTH_PASSWORD"`
	Retry    int    `file:"retry" env:"TELEMETRY_REPORT_STRICT_RETRY" default:"3"`
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
