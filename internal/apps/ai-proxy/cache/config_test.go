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
	"testing"
	"time"
)

func TestIsCacheEnabled(t *testing.T) {
	// save original environment variable
	originalValue := os.Getenv(envCacheEnabled)
	defer func() {
		// restore original value
		if originalValue == "" {
			os.Unsetenv(envCacheEnabled)
		} else {
			os.Setenv(envCacheEnabled, originalValue)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		unset    bool
		expected bool
	}{
		{
			name:     "unset environment variable should enable cache",
			unset:    true,
			expected: true,
		},
		{
			name:     "empty string should enable cache",
			envValue: "",
			expected: true,
		},
		{
			name:     "true should enable cache",
			envValue: "true",
			expected: true,
		},
		{
			name:     "TRUE should enable cache",
			envValue: "TRUE",
			expected: true,
		},
		{
			name:     "1 should enable cache",
			envValue: "1",
			expected: true,
		},
		{
			name:     "false should disable cache",
			envValue: "false",
			expected: false,
		},
		{
			name:     "FALSE should disable cache",
			envValue: "FALSE",
			expected: false,
		},
		{
			name:     "False should disable cache",
			envValue: "False",
			expected: false,
		},
		{
			name:     "random string should enable cache",
			envValue: "random",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unset {
				os.Unsetenv(envCacheEnabled)
			} else {
				os.Setenv(envCacheEnabled, tt.envValue)
			}

			result := isCacheEnabled()
			if result != tt.expected {
				t.Errorf("isCacheEnabled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetRefreshInterval(t *testing.T) {
	// save original environment variable
	originalValue := os.Getenv(envCacheRefreshInterval)
	defer func() {
		// restore original value
		if originalValue == "" {
			os.Unsetenv(envCacheRefreshInterval)
		} else {
			os.Setenv(envCacheRefreshInterval, originalValue)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		unset    bool
		expected time.Duration
	}{
		{
			name:     "unset environment variable should use default 1m",
			unset:    true,
			expected: 1 * time.Minute,
		},
		{
			name:     "empty string should use default 1m",
			envValue: "",
			expected: 1 * time.Minute,
		},
		{
			name:     "30s should parse correctly",
			envValue: "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "2m should parse correctly",
			envValue: "2m",
			expected: 2 * time.Minute,
		},
		{
			name:     "1h should parse correctly",
			envValue: "1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "1h30m should parse correctly",
			envValue: "1h30m",
			expected: 1*time.Hour + 30*time.Minute,
		},
		{
			name:     "invalid format should fallback to default 1m",
			envValue: "invalid",
			expected: 1 * time.Minute,
		},
		{
			name:     "random string should fallback to default 1m",
			envValue: "not-a-time",
			expected: 1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unset {
				os.Unsetenv(envCacheRefreshInterval)
			} else {
				os.Setenv(envCacheRefreshInterval, tt.envValue)
			}

			result := getRefreshInterval()
			if result != tt.expected {
				t.Errorf("getRefreshInterval() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// save original environment variables
	originalEnabled := os.Getenv(envCacheEnabled)
	originalInterval := os.Getenv(envCacheRefreshInterval)
	defer func() {
		// restore original values
		if originalEnabled == "" {
			os.Unsetenv(envCacheEnabled)
		} else {
			os.Setenv(envCacheEnabled, originalEnabled)
		}
		if originalInterval == "" {
			os.Unsetenv(envCacheRefreshInterval)
		} else {
			os.Setenv(envCacheRefreshInterval, originalInterval)
		}
	}()

	tests := []struct {
		name             string
		enabledEnvValue  string
		intervalEnvValue string
		expectedEnabled  bool
		expectedInterval time.Duration
	}{
		{
			name:             "default config - cache enabled, 1m interval",
			enabledEnvValue:  "",
			intervalEnvValue: "",
			expectedEnabled:  true,
			expectedInterval: 1 * time.Minute,
		},
		{
			name:             "cache disabled, custom interval",
			enabledEnvValue:  "false",
			intervalEnvValue: "5m",
			expectedEnabled:  false,
			expectedInterval: 5 * time.Minute,
		},
		{
			name:             "cache enabled, custom interval",
			enabledEnvValue:  "true",
			intervalEnvValue: "2h",
			expectedEnabled:  true,
			expectedInterval: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(envCacheEnabled, tt.enabledEnvValue)
			os.Setenv(envCacheRefreshInterval, tt.intervalEnvValue)

			config := loadConfig()

			if config.Enabled != tt.expectedEnabled {
				t.Errorf("loadConfig().Enabled = %v, expected %v", config.Enabled, tt.expectedEnabled)
			}

			if config.RefreshInterval != tt.expectedInterval {
				t.Errorf("loadConfig().RefreshInterval = %v, expected %v", config.RefreshInterval, tt.expectedInterval)
			}
		})
	}
}
