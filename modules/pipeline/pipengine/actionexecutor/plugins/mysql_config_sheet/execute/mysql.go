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

package execute

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/erda-project/erda/pkg/clusterdialer"
)

type Request struct {
	Host       string
	Port       string
	User       string
	Password   string
	ClusterKey string
}

func (r Request) dsn(proto string) string {
	return fmt.Sprintf("%s:%s@%s(%s:%s)/?charset=utf8mb4,utf8&parseTime=true&multiStatements=true", r.User, r.Password, proto, r.Host, r.Port)
}

func (r Request) dbOpen() (*sql.DB, error) {
	proto := "tcp"
	proto = fmt.Sprintf("tcp-%s", r.ClusterKey)

	mysql.RegisterDialContext(proto, mysql.DialContextFunc(clusterdialer.DialContextTCP(r.ClusterKey)))
	db, err := sql.Open("mysql", r.dsn(proto))
	if err != nil {
		return nil, err
	}
	return db, nil
}
