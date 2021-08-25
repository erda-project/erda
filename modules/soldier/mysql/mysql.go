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

package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
)

type Request struct {
	Sqls      []string `json:"sqls"`
	Host      string   `json:"host"`
	Url       string   `json:"url"`
	User      string   `json:"user"`
	Password  string   `json:"password"`
	OssUrl    string   `json:"ossUrl"`
	CreateDbs []string `json:"createDbs"`
}

func (r Request) Addr() (string, string) {
	host, port, _ := net.SplitHostPort(strings.Replace(r.Url, "jdbc:mysql://", "", -1))
	return host, port
}

func (r Request) DSN() string {
	host, port := r.Addr()
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4,utf8&parseTime=true", r.User, r.Password, host, port)
}

func (r Request) Exec(q string) error {
	db, err := sql.Open("mysql", r.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cf := context.WithTimeout(context.Background(), time.Minute)
	defer cf()
	_, err = db.ExecContext(ctx, q)
	return err
}

func Init(w http.ResponseWriter, r *http.Request) {
	var a []Request
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	logrus.Infoln("mysql init:", a)
	for _, b := range a {
		for _, q := range b.Sqls {
			logrus.Infof("init sql:%v", string(q))
			if err := b.Exec(q); err != nil {
				logrus.Errorln(err)
				colonyutil.WriteErr(w, "1000", err.Error())
				return
			}
		}
	}
	colonyutil.WriteData(w, true)
}

func Check(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	logrus.Infoln("mysql check:", req)
	db, err := sql.Open("mysql", req.DSN())
	if err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "1000", err.Error())
		return
	}
	defer db.Close()
	ctx, cf := context.WithTimeout(context.Background(), time.Minute)
	defer cf()
	rows, err := db.QueryContext(ctx, "show slave status")
	if err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "1000", err.Error())
		return
	}
	defer rows.Close()
	if !rows.Next() {
		colonyutil.WriteErr(w, "1000", "no rows")
		return
	}
	cols, err := rows.Columns()
	if err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "1000", err.Error())
		return
	}
	a := make([]interface{}, len(cols))
	for i := range a {
		if cols[i] == "Slave_IO_Running" {
			a[i] = new(string)
		} else if cols[i] == "Slave_SQL_Running" {
			a[i] = new(string)
		} else {
			a[i] = new(interface{})
		}
	}
	err = rows.Scan(a...)
	if err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "1000", err.Error())
		return
	}
	m := make(map[string]interface{}, 2)
	for i, col := range cols {
		if col == "Slave_IO_Running" {
			m["slaveIoRunning"] = *(a[i].(*string))
		} else if col == "Slave_SQL_Running" {
			m["slaveSqlRunning"] = *(a[i].(*string))
		}
	}
	colonyutil.WriteData(w, m)
}

func Exec(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	logrus.Infoln("mysql exec:", req)
	for _, q := range req.Sqls {
		if err := req.Exec(q); err != nil {
			logrus.Errorln(err)
			colonyutil.WriteErr(w, "1000", err.Error())
			return
		}
	}
	colonyutil.WriteData(w, true)
}

func ExecFile(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	logrus.Infoln("mysql exec file:", req)
	ctx, cf := context.WithTimeout(context.Background(), time.Minute)
	defer cf()
	host, port := req.Addr()
	c := exec.CommandContext(ctx, "bash",
		"/app/sql.sh", host, port, req.User, req.Password, strings.Join(req.CreateDbs, "\n"), req.OssUrl)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = []string{"NETDATA_SQLDATA_PATH=" + os.Getenv("NETDATA_SQLDATA_PATH")}
	if err := c.Run(); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "1000", err.Error())
		return
	}
	colonyutil.WriteData(w, true)
}
