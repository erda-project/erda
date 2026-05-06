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

package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalConfigResolvedHost(t *testing.T) {
	require.Equal(t, "https://host.example.com", (&GlobalConfig{Host: "https://host.example.com"}).ResolvedHost())
	require.Equal(t, "https://server.example.com", (&GlobalConfig{Server: "https://server.example.com"}).ResolvedHost())
	require.Empty(t, (&GlobalConfig{}).ResolvedHost())
}

func TestSetAndGetGlobalConfigFromFile(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config")
	expected := &GlobalConfig{
		Version:       ConfigVersion,
		Host:          "https://host.example.com",
		UpdateChannel: "alpha",
	}

	err := SetGlobalConfig(configFile, expected)
	require.NoError(t, err)

	actual, err := GetGlobalConfigFrom(configFile)
	require.NoError(t, err)
	require.Equal(t, expected.Host, actual.Host)
	require.Equal(t, expected.Version, actual.Version)
	require.Equal(t, expected.UpdateChannel, actual.UpdateChannel)
}

func TestGetGlobalConfigFromSupportsLegacyServerField(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config")
	err := os.WriteFile(configFile, []byte("version: v0.0.1\nserver: https://legacy.example.com\n"), 0o644)
	require.NoError(t, err)

	actual, err := GetGlobalConfigFrom(configFile)
	require.NoError(t, err)
	require.Equal(t, "https://legacy.example.com", actual.ResolvedHost())
}
