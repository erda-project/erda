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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains all test configuration
type Config struct {
	// Basic configuration
	Host    string
	Token   string
	Timeout time.Duration

	// MCP Server test configuration for proxy tests
	TestMCPName string
	TestMCPTag  string

	TestToolName  string
	TestArgsName  string
	TestArgsValue string
}

var globalConfig *Config

// Load loads configuration from .env file
func Load() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	config := &Config{
		Host:        "http://localhost:8081",
		Token:       "",
		Timeout:     2 * time.Minute,
		TestMCPName: "fetch",
		TestMCPTag:  "1.0.0",
	}

	// Load configuration from .env file
	if err := loadEnvFile(config); err != nil {
		fmt.Printf("Warning: Failed to load .env file: %v\n", err)
		fmt.Println("Using default configuration values")
	}

	globalConfig = config
	return config
}

// Get returns global configuration
func Get() *Config {
	if globalConfig == nil {
		return Load()
	}
	return globalConfig
}

// loadEnvFile loads configuration from .env file
func loadEnvFile(config *Config) error {
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle key=value format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				setConfigValue(config, key, value)
			}
		}
	}

	return scanner.Err()
}

// setConfigValue sets configuration value based on key
func setConfigValue(config *Config, key, value string) {
	switch key {
	case "HOST":
		config.Host = value
	case "TOKEN":
		config.Token = value
	case "TIMEOUT":
		if timeout, err := strconv.Atoi(value); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}
	case "TEST_MCP_NAME":
		config.TestMCPName = value
	case "TEST_MCP_TAG":
		config.TestMCPTag = value
	case "TEST_TOOL_NAME":
		config.TestToolName = value
	case "TEST_ARGS_NAME":
		config.TestArgsName = value
	case "TEST_ARGS_VALUE":
		config.TestArgsValue = value
	}
}
