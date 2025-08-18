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

	// Test models categorized by API type
	ChatModels       []string
	EmbeddingsModels []string
	AudioModels      []string
	ImageModels      []string
	ResponsesModels  []string
	ThinkingModels   []string
}

var globalConfig *Config

// Load loads configuration from .env file
func Load() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	config := &Config{
		Host:    "http://localhost:8081",
		Token:   "",
		Timeout: 2 * time.Minute, // 2 minutes, consistent with production gateway timeout
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

// setConfigValue sets configuration value
func setConfigValue(config *Config, key, value string) {
	// Use environment variable if exists, otherwise use .env file value
	if envValue := os.Getenv(key); envValue != "" {
		value = envValue
	}

	switch key {
	case "AI_PROXY_HOST":
		config.Host = value
	case "AI_PROXY_TOKEN":
		config.Token = value
	case "TEST_TIMEOUT":
		if timeout, err := strconv.Atoi(value); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}
	case "CHAT_MODELS":
		config.ChatModels = parseModelList(value)
	case "EMBEDDINGS_MODELS":
		config.EmbeddingsModels = parseModelList(value)
	case "AUDIO_MODELS":
		config.AudioModels = parseModelList(value)
	case "IMAGE_MODELS":
		config.ImageModels = parseModelList(value)
	case "RESPONSES_MODELS":
		config.ResponsesModels = parseModelList(value)
	case "THINKING_MODELS":
		config.ThinkingModels = parseModelList(value)
	}
}

// parseModelList parses model list
func parseModelList(value string) []string {
	if value == "" {
		return nil
	}

	var models []string
	for _, model := range strings.Split(value, ",") {
		model = strings.TrimSpace(model)
		if model != "" {
			models = append(models, model)
		}
	}
	return models
}

// GetAllModels returns all configured models
func (c *Config) GetAllModels() []string {
	var allModels []string
	allModels = append(allModels, c.ChatModels...)
	allModels = append(allModels, c.EmbeddingsModels...)
	allModels = append(allModels, c.AudioModels...)
	allModels = append(allModels, c.ImageModels...)
	allModels = append(allModels, c.ResponsesModels...)
	return allModels
}
