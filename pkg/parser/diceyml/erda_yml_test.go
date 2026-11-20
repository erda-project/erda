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

package diceyml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErdaYmlMySQLTLSDefaults(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "erda.yml"))
	require.NoError(t, err)

	yml, err := New(data, false)
	require.NoError(t, err)

	obj := yml.Obj()
	require.NotContains(t, obj.Envs, "MYSQL_TLS")
	require.NotContains(t, obj.Envs, "MYSQL_CACERTPATH")
	for _, name := range []string{
		"erda-server", "cluster-manager", "gittar", "hepa", "monitor",
		"orchestrator", "pipeline", "streaming", "log-service",
	} {
		service := obj.Services[name]
		require.NotNil(t, service, name)
		require.Empty(t, service.Envs["MYSQL_TLS"], name)
		require.Empty(t, service.Envs["MYSQL_CACERTPATH"], name)
	}
	for _, name := range []string{"cluster-agent", "collector", "collector-agent"} {
		service := obj.Services[name]
		require.NotNil(t, service, name)
		require.NotContains(t, service.Envs, "MYSQL_TLS", name)
		require.NotContains(t, service.Envs, "MYSQL_CACERTPATH", name)
	}
}
