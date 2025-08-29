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

package cache

import (
	"os"
	"strings"
	"time"
)

const (
	// environment variable keys
	envCacheEnabled         = "AI_PROXY_CACHE_ENABLED"
	envCacheRefreshInterval = "AI_PROXY_CACHE_REFRESH_INTERVAL"
)

// config holds all cache-related configuration
type config struct {
	Enabled         bool
	RefreshInterval time.Duration
}

// loadConfig loads cache configuration from environment variables
func loadConfig() *config {
	return &config{
		Enabled:         isCacheEnabled(),
		RefreshInterval: getRefreshInterval(),
	}
}

// isCacheEnabled checks if cache is enabled via environment variable
// Set AI_PROXY_CACHE_ENABLED=false to disable cache
func isCacheEnabled() bool {
	enabled := os.Getenv(envCacheEnabled)
	return enabled == "" || strings.ToLower(enabled) != "false"
}

// getRefreshInterval gets refresh interval from environment variable
// Set AI_PROXY_CACHE_REFRESH_INTERVAL=2m (default: 1m)
// Supports time formats like: 30s, 1m, 5m, 1h, 1h30m
func getRefreshInterval() time.Duration {
	interval := os.Getenv(envCacheRefreshInterval)
	if interval != "" {
		if duration, err := time.ParseDuration(interval); err == nil {
			return duration
		}
	}

	// fallback to default if parsing fails
	return 1 * time.Minute
}
