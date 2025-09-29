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

package config

import (
	"embed"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type DiceInfo struct {
	LocalClusterName string `file:"local_cluster_name" env:"DICE_CLUSTER_NAME"`
	Namespace        string `file:"namespace" env:"DICE_NAMESPACE"`
}

type McpScanConfig struct {
	Enable                    bool          `file:"enable" env:"MCP_SCAN_ENABLE"`
	McpClusters               string        `file:"mcp_clusters" default:"" env:"MCP_CLUSTERS"`
	SyncClusterConfigInterval time.Duration `file:"sync_cluster_config_interval" default:"10m" env:"SYNC_CLUSTER_CONFIG_INTERVAL"`
}

type Config struct {
	LogLevelStr string       `file:"log_level" default:"info" env:"LOG_LEVEL"`
	LogLevel    logrus.Level `json:"-" yaml:"-"`

	SelfURL           string `file:"self_url" env:"SELF_URL" required:"true"`
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"MCP_PROXY_PUBLIC_URL"`

	IsMcpProxy    bool          `file:"is_mcp_proxy" env:"IS_MCP_PROXY"`
	McpScanConfig McpScanConfig `file:"mcp_scan_config"`
	DiceInfo      DiceInfo      `file:"dice_info"`

	McpClusters               string        `file:"mcp_clusters" default:"" env:"MCP_CLUSTERS"`
	SyncClusterConfigInterval time.Duration `file:"sync_cluster_config_interval" default:"10m" env:"SYNC_CLUSTER_CONFIG_INTERVAL"`

	EmbedRoutesFS embed.FS
}

var embedRoutesFS embed.FS

func InjectEmbedRoutesFS(in embed.FS) {
	embedRoutesFS = in
}

// DoPost do some post process after config loaded
func (cfg *Config) DoPost() error {
	// routes fs
	cfg.EmbedRoutesFS = embedRoutesFS

	// parse log level
	level, err := logrus.ParseLevel(cfg.LogLevelStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level, level: %s, err: %v", cfg.LogLevel, err)
	}
	logrus.SetLevel(level)
	cfg.LogLevel = level

	return nil
}
