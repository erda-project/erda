// Copyright (c) 2023 Terminus, Inc.
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

package mysql

import (
	"context"
	"github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/pkg/mysqlhelper"
	"strings"
)

// InstanceAdapter is the adapter to execute MySQL command for MySQL instance.

var (
	_ InstanceAdapter = (*localClient)(nil)
	_ InstanceAdapter = (*operatorClient)(nil)
)

type InstanceAdapter interface {
	ExecSQLs() error
}

func NewInstanceAdapter(options ...options) InstanceAdapter {
	var o = option{
		useOperator:    false,
		basicConfig:    new(pb.OperatorInstanceBasicConfig),
		clusterConfig:  make(map[string]string),
		operatorInsCli: nil,
	}
	for _, f := range options {
		f(&o)
	}
	if o.useOperator {
		return &operatorClient{o: &o}
	}
	return &localClient{o: &o}
}

type localClient struct {
	o *option
}

func (l *localClient) ExecSQLs() error {
	return (&mysqlhelper.Request{
		ClusterKey: l.o.basicConfig.GetClusterKey(),
		Url:        "jdbc:mysql://" + l.o.basicConfig.GetWAddress(),
		User:       l.o.basicConfig.GetUsername(),
		Password:   l.o.basicConfig.GetPassword(),
		Sqls:       l.o.queries,
	}).Exec()
}

type operatorClient struct {
	o *option
}

func (o *operatorClient) ExecSQLs() error {
	_, err := o.o.operatorInsCli.ExecSQL(context.Background(), &pb.ExecSQLRequest{
		Config:    o.o.basicConfig,
		QueryType: pb.QueryType_sql,
		Query:     strings.Join(o.o.queries, "\n"),
	})
	return err
}

type option struct {
	useOperator    bool
	basicConfig    *pb.OperatorInstanceBasicConfig
	clusterConfig  map[string]string
	operatorInsCli pb.MySQLOperatorInstanceServiceServer
	queries        []string
}

type options func(o *option)

func withUseOperator(useOperator bool) options {
	return func(o *option) {
		o.useOperator = useOperator
	}
}

func withUsername(username string) options {
	return func(o *option) {
		o.basicConfig.Username = username
	}
}

func withPassword(password string) options {
	return func(o *option) {
		o.basicConfig.Password = password
	}
}

func withSchema(schema string) options {
	return func(o *option) {
		o.basicConfig.Schema = schema
	}
}

func withClusterKey(clusterKey string) options {
	return func(o *option) {
		o.basicConfig.ClusterKey = clusterKey
	}
}

func withAddress(address string) options {
	return func(o *option) {
		o.basicConfig.WAddress = address
	}
}

func withQueries(queries []string) options {
	return func(o *option) {
		o.queries = queries
	}
}
