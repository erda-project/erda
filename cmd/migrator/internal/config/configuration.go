// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/sqllint/linters"
	"github.com/erda-project/erda/pkg/sqllint/rules"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	versionPackage  = "/opt/dice-tools/versionpackage"
	versionFilename = versionPackage + "/version"
)

var configuration *Configuration

type Configuration struct {
	envs *envs
	cf   *ConfigFile
}

func Config() *Configuration {
	if configuration == nil {
		configuration = new(Configuration)
		if err := configuration.reload(); err != nil {
			log.Errorf("failed to reload configuration: %v", err)
		}

		log.Infof("%+v", *configuration.envs)
	}
	return configuration
}

func (c Configuration) DSN() string {
	if c.envs.MySQLUsername != "" {
		return fmt.Sprintf("%s:%s@(%s:%v)/",
			c.envs.MySQLUsername,
			c.envs.MySQLPassword,
			c.envs.MySQLHost,
			c.envs.MySQLPort,
		)
	}

	if c.cf == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s@(%s:%v)/",
		c.cf.Installs.Addons.Mysql.User,
		c.cf.Installs.Addons.Mysql.Password,
		c.cf.Installs.Addons.Mysql.Host,
		c.cf.Installs.Addons.Mysql.Port,
	)
}

func (c Configuration) SandboxDSN() string {
	return fmt.Sprintf("root:%s@(localhost:3306)/", c.envs.SandboxRootPassword)
}

func (c Configuration) Database() string {
	if c.envs.MySQlDBName != "" {
		return c.envs.MySQlDBName
	}

	if c.cf == nil {
		return ""
	}

	return c.cf.Installs.Addons.Mysql.Db
}

// MigrationDir returns migrations scripts dir
func (c Configuration) MigrationDir() string {
	if c.envs.MigrationDir != "" {
		return c.envs.MigrationDir
	}

	data, err := ioutil.ReadFile(versionFilename)
	if err != nil {
		return ""
	}
	migrationDir := filepath.Join(versionPackage, string(data))
	return migrationDir
}

// Workdir returns workdir to join the scripts' path
func (c Configuration) Workdir() string {
	return ""
}

// DebugSQL returns whether  the process need to debug SQLs
func (c Configuration) DebugSQL() bool {
	return c.envs.DebugSQL
}

// NeedErdaMySQLLint returns whether the process need to lint the SQLs
func (c Configuration) NeedErdaMySQLLint() bool {
	return c.envs.ErdaLint
}

// Modules returns the modules for installing
func (c Configuration) Modules() []string {
	return strings.Split(c.envs.Modules_, ",")
}

// Rules returns Erda MySQL linters
// note: hard code here
func (c Configuration) Rules() []rules.Ruler {
	var ddls = linters.CreateTableStmt | linters.CreateIndexStmt | linters.DropIndexStmt |
		linters.AlterTableOption | linters.AlterTableAddColumns | linters.AlterTableAddConstraint |
		linters.AlterTableDropIndex | linters.AlterTableModifyColumn | linters.AlterTableModifyColumn |
		linters.AlterTableChangeColumn | linters.AlterTableAlterColumn | linters.AlterTableRenameIndex

	var dmls = linters.SelectStmt | linters.UnionStmt | linters.LoadDataStmt |
		linters.InsertStmt | linters.DeleteStmt | linters.UpdateStmt |
		linters.ShowStmt | linters.SplitRegionStmt

	list := []rules.Ruler{
		linters.NewBooleanFieldLinter,
		linters.NewCharsetLinter,
		linters.NewColumnNameLinter,
		linters.NewColumnCommentLinter,
		linters.NewFloatDoubleLinter,
		linters.NewForeignKeyLinter,
		linters.NewIndexLengthLinter,
		linters.NewIndexNameLinter,
		linters.NewKeywordsLinter,
		linters.NewIDExistsLinter,
		linters.NewIDTypeLinter,
		linters.NewIDIsPrimaryLinter,
		linters.NewCreatedAtExistsLinter,
		linters.NewCreatedAtDefaultValueLinter,
		linters.NewUpdatedAtExistsLinter,
		linters.NewUpdatedAtTypeLinter,
		linters.NewUpdatedAtDefaultValueLinter,
		linters.NewUpdatedAtOnUpdateLinter,
		linters.NewNotNullLinter,
		linters.NewTableCommentLinter,
		linters.NewTableNameLinter,
		linters.NewVarcharLengthLinter,
		linters.NewAllowedStmtLinter(ddls, dmls),
	}

	return list
}

// reload reloads the envs and ${DICE_CONFIG}/config.yaml
func (c *Configuration) reload() error {
	c.envs = new(envs)
	if err := envconf.Load(c.envs); err != nil {
		return errors.Wrap(err, "failed to Load envs")
	}

	if data, err := ioutil.ReadFile(c.envs.ConfigPath); err == nil {
		c.cf = new(ConfigFile)
		_ = yaml.Unmarshal(data, c.cf) // allows err
	}

	return nil
}

type envs struct {
	ConfigPath string `env:"CONFIGPATH"` // ${DICE_CONFIG}/config.yaml
	DiceConfig string `env:"DICE_CONFIG"`

	MySQLHost     string `env:"MIGRATION_MYSQL_HOST"`
	MySQLPort     uint64 `env:"MIGRATION_MYSQL_PORT"`
	MySQLUsername string `env:"MIGRATION_MYSQL_USERNAME"`
	MySQLPassword string `env:"MIGRATION_MYSQL_PASSWORD"`
	MySQlDBName   string `env:"MIGRATION_MYSQL_DBNAME"`
	DebugSQL      bool   `env:"MIGRATION_DEBUGSQL"`
	ErdaLint      bool   `env:"MIGRATION_ERDA_LINT"`
	Modules_      string `env:"MIGRATION_MODULES"`

	SandboxRootPassword string `env:"MYSQL_ROOT_PASSWORD"`

	Workdir      string `env:"WORKDIR"`
	MigrationDir string `env:"MIGRATION_DIR"`
}

// ConfigFile represents the structure of ${DICE_CONFIG}/config.yaml .
// can read mysql configurations from this.
type ConfigFile struct {
	Version  string `json:"version" yaml:"version"`
	Installs struct {
		DataDir    string `json:"data_dir" yaml:"data_dir"`
		NetdataDir string `json:"netdata_dir" yaml:"netdata_dir"`
		Addons     struct {
			Mysql struct {
				Host     string `json:"host" yaml:"host"`
				Port     int    `json:"port" yaml:"port"`
				User     string `json:"user" yaml:"user"`
				Password string `json:"password" yaml:"password"`
				Db       string `json:"db" yaml:"db"`
			} `json:"mysql" yaml:"mysql"`
		} `json:"addons" yaml:"addons"`
	} `json:"installs" yaml:"installs"`
}
