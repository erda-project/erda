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

package migrator

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

const (
	smallestRetryTimeout = 60
	certName             = "custom"
)

type Parameters interface {
	ScriptsParameters

	// MySQLParameters returns MySQL DSN configuration
	MySQLParameters() *DSNParameters

	// SandboxParameters returns Sandbox DSN configuration
	SandboxParameters() *DSNParameters

	// DebugSQL gets weather to debug SQL executing
	DebugSQL() bool

	// RetryTimeout unit: second
	RetryTimeout() uint64

	SkipMigrationLint() bool
	SkipSandbox() bool
	SkipPreMigrate() bool
	SkipMigrate() bool
}

type DSNParameters struct {
	*TLSConfig
	Username  string
	Password  string
	Host      string
	Port      int
	Database  string
	ParseTime bool
	TLS       string
	Timeout   time.Duration
}
type TLSConfig struct {
	DBClientKey  string
	DBCaCert     string
	DBClientCert string
}

func (c DSNParameters) Format(database bool) (dsn string) {
	if c.TLS == certName {
		if c.DBCaCert == "" {
			logrus.Error("MysqlCaCert can not be empty.")
			return ""
		}
		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM([]byte(c.DBCaCert)); !ok {
			logrus.Error("failed to append ca cert.")
			return ""
		}
		clientCert := make([]tls.Certificate, 0)
		if c.DBClientCert != "" && c.DBClientKey != "" {
			certs, err := tls.LoadX509KeyPair(c.DBClientCert, c.DBClientKey)
			if err != nil {
				logrus.Error(err)
				return ""
			}
			clientCert = append(clientCert, certs)
		}
		err := mysql.RegisterTLSConfig(c.TLS, &tls.Config{
			RootCAs:      rootCertPool,
			Certificates: clientCert,
		})
		if err != nil {
			logrus.Error(err)
			return ""
		}
	}
	var dbName string
	if database {
		dbName = c.Database
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 150 * time.Second
	}
	mc := mysql.Config{
		User:              c.Username,
		Passwd:            c.Password,
		Addr:              fmt.Sprintf("%s:%d", c.Host, c.Port),
		DBName:            dbName,
		Net:               "tcp",
		ParseTime:         c.ParseTime,
		Timeout:           timeout,
		MultiStatements:   true,
		CheckConnLiveness: true,
		Params: map[string]string{
			"charset": "utf8mb4,utf8",
		},
		AllowNativePasswords: true,
		TLSConfig:            c.TLS,
		Loc:                  time.Local,
	}

	return mc.FormatDSN()
}

type SQLCollectorDir interface {
	SQLCollectorDir() string
}

func SQLCollectorFilename() string {
	return fmt.Sprintf("%s_erda-migrator.sql.log", time.Now().Format("20060102150405"))
}

func RetryTimeout(timeout uint64) time.Duration {
	if timeout > smallestRetryTimeout {
		return time.Second * time.Duration(timeout)
	}
	return time.Second * smallestRetryTimeout
}
