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

package dao

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/core/legacy/conf"
)

func TestMySQLDSN(t *testing.T) {
	tests := []struct {
		name string
		tls  string
		want string
	}{
		{
			name: "without TLS",
			want: "erda:placeholder@tcp(mysql:3306)/erda?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "with named TLS",
			tls:  "rds",
			want: "erda:placeholder@tcp(mysql:3306)/erda?charset=utf8mb4&parseTime=True&loc=Local&tls=rds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mysqlDSN("erda", "placeholder", "mysql", "3306", "erda", "Local", tt.tls)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogDBInitializationDoesNotLogPassword(t *testing.T) {
	const marker = "credential-marker-must-not-be-logged"
	t.Setenv("MYSQL_HOST", "mysql")
	t.Setenv("MYSQL_PORT", "3306")
	t.Setenv("MYSQL_USERNAME", "erda")
	t.Setenv("MYSQL_PASSWORD", marker)
	t.Setenv("MYSQL_DATABASE", "erda")
	t.Setenv("MYSQL_TLS", "")
	conf.LoadForTest()

	var output bytes.Buffer
	oldOutput := logrus.StandardLogger().Out
	oldLevel := logrus.GetLevel()
	logrus.SetOutput(&output)
	logrus.SetLevel(logrus.DebugLevel)
	defer func() {
		logrus.SetOutput(oldOutput)
		logrus.SetLevel(oldLevel)
	}()

	logDBInitialization()
	assert.False(t, strings.Contains(output.String(), marker), output.String())
}
