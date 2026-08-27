package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParametersUseRDSConfiguration(t *testing.T) {
	t.Setenv("MYSQL_HOST", "rds.example")
	t.Setenv("MYSQL_PORT", "3306")
	t.Setenv("MYSQL_USER", "erda_test")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DICE_DB", "erda")
	t.Setenv("MYSQL_TLS", "rds")
	t.Setenv("MIGRATION_MODULES", "pipeline, qa")

	p := parameters{}
	require.Equal(t, "rds.example", p.MySQLParameters().Host)
	require.Equal(t, "rds", p.MySQLParameters().TLS)
	require.Equal(t, "erda", p.MySQLParameters().Database)
	require.Equal(t, []string{"pipeline", "qa"}, p.Modules())
}

func TestParametersDefaultToLegacyPackageDirectory(t *testing.T) {
	t.Setenv("MIGRATION_DIR", "")
	require.Equal(t, "/opt/dice-tools/versionpackage/sqls/erda", (parameters{}).MigrationDir())
}
