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

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParametersUseRDSConfiguration(t *testing.T) {
	t.Setenv("MYSQL_HOST", "mysql.example")
	t.Setenv("MYSQL_PORT", "3306")
	t.Setenv("MYSQL_USER", "erda_test")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DICE_DB", "erda")
	t.Setenv("MYSQL_TLS", "mysql")
	t.Setenv("MIGRATION_MODULES", "pipeline, qa")

	p := parameters{}
	require.Equal(t, "mysql.example", p.MySQLParameters().Host)
	require.Equal(t, "mysql", p.MySQLParameters().TLS)
	require.Equal(t, "erda", p.MySQLParameters().Database)
	require.Equal(t, []string{"pipeline", "qa"}, p.Modules())
}

func TestParametersDefaultToLegacyPackageDirectory(t *testing.T) {
	t.Setenv("MIGRATION_DIR", "")
	require.Equal(t, "/opt/dice-tools/versionpackage/sqls/erda", (parameters{}).MigrationDir())
}
