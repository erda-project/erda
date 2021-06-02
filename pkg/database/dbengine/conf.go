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

package dbengine

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/pkg/envconf"
)

type Conf struct {
	MySQLURL          string `env:"MYSQL_URL"`
	MySQLHost         string `env:"MYSQL_HOST"`
	MySQLPort         string `env:"MYSQL_PORT"`
	MySQLUsername     string `env:"MYSQL_USERNAME"`
	MySQLPassword     string `env:"MYSQL_PASSWORD"`
	MySQLDatabase     string `env:"MYSQL_DATABASE"`
	MySQLCharset      string `env:"MYSQL_CHARSET" default:"utf8mb4"`
	MySQLMaxIdleConns int    `env:"MYSQL_MAXIDLECONNS"`
	MySQLMaxOpenConns int    `env:"MYSQL_MAXOPENCONNS"`
	MySQLMaxLifeTime  int64  `env:"MYSQL_MAXLIFETIME"` // 单位秒 (s)
	Debug             bool   `env:"DEBUG"`
}

func LoadDefaultConf() (*Conf, error) {
	cfg := Conf{}
	if err := envconf.Load(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// url 返回 MySQL 连接地址
func (cfg *Conf) url() string {
	if cfg.MySQLURL != "" {
		return cfg.MySQLURL
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQLUsername, cfg.MySQLPassword, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDatabase, cfg.MySQLCharset)
}

// maxIdleConns 返回 MySQL 最大连接池
func (cfg *Conf) maxIdleConns() int {
	if cfg.MySQLMaxIdleConns > 0 {
		return cfg.MySQLMaxIdleConns
	}
	return 5
}

// maxOpenConns 返回 MySQL 最大连接数
func (cfg *Conf) maxOpenConns() int {
	if cfg.MySQLMaxOpenConns > 0 {
		return cfg.MySQLMaxOpenConns
	}
	return 50
}

// maxLifeTime 返回 MySQL 连接最大存活时间
func (cfg *Conf) maxLifeTime() time.Duration {
	if cfg.MySQLMaxLifeTime > 0 {
		return time.Duration(cfg.MySQLMaxLifeTime) * time.Second
	}
	return 180 * time.Second
}
