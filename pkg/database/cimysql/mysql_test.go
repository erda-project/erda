// Copyright (c) 2026 Terminus, Inc.
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

package cimysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineIsNotCreatedDuringPackageInitialization(t *testing.T) {
	assert.Nil(t, Engine)
}

func TestConfigDSN(t *testing.T) {
	tests := []struct {
		name string
		tls  string
		want string
	}{
		{
			name: "without TLS",
			want: "erda:secret@tcp(mysql:3306)/erda?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "with named TLS",
			tls:  "rds",
			want: "erda:secret@tcp(mysql:3306)/erda?charset=utf8mb4&parseTime=True&loc=Local&tls=rds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config{
				Host:     "mysql",
				Port:     3306,
				Username: "erda",
				Password: "secret",
				Database: "erda",
				Charset:  "utf8mb4",
				TLS:      tt.tls,
			}
			assert.Equal(t, tt.want, cfg.dsn())
		})
	}
}

func TestInitialize(t *testing.T) {
	t.Setenv("MYSQL_HOST", "mysql")
	t.Setenv("MYSQL_PORT", "3306")
	t.Setenv("MYSQL_USERNAME", "erda")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DATABASE", "erda")
	t.Setenv("MYSQL_TLS", "")

	assert.NoError(t, initialize())
	if assert.NotNil(t, Engine) {
		assert.Equal(t, "erda:secret@tcp(mysql:3306)/erda?charset=utf8mb4&parseTime=True&loc=Local", Engine.DataSourceName())
		assert.NoError(t, Engine.Close())
	}
	Engine = nil
}

func TestInitializeRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("MYSQL_PORT", "not-a-port")

	assert.Error(t, initialize())
}
