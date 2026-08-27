package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/database/mysqltls"
	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqlparser/migrator"
)

type parameters struct{}

func (parameters) Workdir() string { return "" }

func (parameters) MigrationDir() string {
	if dir := os.Getenv("MIGRATION_DIR"); dir != "" {
		return dir
	}
	return "/opt/dice-tools/versionpackage/sqls/erda"
}

func (parameters) Modules() []string {
	var modules []string
	for _, module := range strings.Split(os.Getenv("MIGRATION_MODULES"), ",") {
		if module = strings.TrimSpace(module); module != "" {
			modules = append(modules, module)
		}
	}
	return modules
}

func (parameters) LintConfig() map[string]sqllint.Config { return nil }

func (parameters) MySQLParameters() *migrator.DSNParameters {
	port, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	return &migrator.DSNParameters{
		Username:  getenvFirst("MYSQL_USER", "MYSQL_USERNAME"),
		Password:  os.Getenv("MYSQL_PASSWORD"),
		Host:      os.Getenv("MYSQL_HOST"),
		Port:      port,
		Database:  getenvFirst("MYSQL_DICE_DB", "MYSQL_DATABASE"),
		ParseTime: true,
		Timeout:   150 * time.Second,
		TLS:       os.Getenv("MYSQL_TLS"),
	}
}

func (parameters) SandboxParameters() *migrator.DSNParameters {
	port, _ := strconv.Atoi(os.Getenv("SANDBOX_MYSQL_PORT"))
	if port == 0 {
		port = 3306
	}
	return &migrator.DSNParameters{
		Username:  "root",
		Password:  getenvFirst("SANDBOX_MYSQL_ROOT_PASSWORD", "SANDBOX_MYSQL_PASSWORD"),
		Host:      getenvFirst("SANDBOX_MYSQL_HOST", "MIGRATION_SANDBOX_HOST", "127.0.0.1"),
		Port:      port,
		Database:  getenvFirst("MYSQL_DICE_DB", "MYSQL_DATABASE"),
		ParseTime: true,
		Timeout:   150 * time.Second,
	}
}

func (parameters) DebugSQL() bool          { return os.Getenv("MIGRATION_DEBUGSQL") == "true" }
func (parameters) RetryTimeout() uint64    { return 150 }
func (parameters) SkipMigrationLint() bool { return envBool("MIGRATION_SKIP_LINT") }
func (parameters) SkipSandbox() bool       { return envBool("MIGRATION_SKIP_SANDBOX") }
func (parameters) SkipPreMigrate() bool    { return envBool("MIGRATION_SKIP_PRE_MIGRATION") }
func (parameters) SkipMigrate() bool       { return envBool("MIGRATION_SKIP_MIGRATION") }

func envBool(name string) bool { return os.Getenv(name) == "true" }

func getenvDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func getenvFirst(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func main() {
	if err := mysqltls.Open(os.Getenv("MYSQL_TLS"), os.Getenv("MYSQL_CACERTPATH"), "", ""); err != nil {
		fmt.Fprintf(os.Stderr, "configure MySQL TLS: %v\n", err)
		os.Exit(1)
	}
	mig, err := migrator.New(parameters{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize migration: %v\n", err)
		os.Exit(1)
	}
	if err := mig.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run migration: %v\n", err)
		os.Exit(1)
	}
}
