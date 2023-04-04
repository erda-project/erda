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
	"net"

	"github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/pkg/mysqlhelper"
)

// InstanceAdapter is the adapter to execute MySQL command for MySQL instance.

var (
	_ InstanceAdapter = (*localClient)(nil)
	_ InstanceAdapter = (*operatorClient)(nil)
)

type InstanceAdapter interface {
	ExecSQLs() error
}

func NewInstanceAdapter(options ...Option) InstanceAdapter {
	var cia = commonInstanceAdapter{
		useOperator:    false,
		execSQLRequest: new(pb.ExecSQLRequest),
		clusterConfig:  make(map[string]string),
		operatorInsCli: nil,
	}
	for _, f := range options {
		f(&cia)
	}
	if cia.useOperator {
		host, _, _ := net.SplitHostPort(cia.execSQLRequest.WAddress)
		cia.execSQLRequest.WAddress = host + ":33080" // 33080 is the mylet port // todo: hard coed here
		return &operatorClient{cia: &cia}
	}
	return &localClient{cia: &cia}
}

type localClient struct {
	cia *commonInstanceAdapter
}

func (l *localClient) ExecSQLs() error {
	return (&mysqlhelper.Request{
		ClusterKey: l.cia.execSQLRequest.GetClusterKey(),
		Url:        "jdbc:mysql://" + l.cia.execSQLRequest.GetWAddress(),
		User:       l.cia.execSQLRequest.GetUsername(),
		Password:   l.cia.execSQLRequest.GetPassword(),
		Sqls:       l.cia.execSQLRequest.GetQueries(),
	}).Exec()
}

type operatorClient struct {
	cia *commonInstanceAdapter
}

func (o *operatorClient) ExecSQLs() error {
	o.cia.execSQLRequest.QueryType = pb.QueryType_sql // only support pb.QueryType_sql yet
	_, err := o.cia.operatorInsCli.ExecSQL(context.Background(), o.cia.execSQLRequest)
	return err
}

type commonInstanceAdapter struct {
	useOperator    bool
	execSQLRequest *pb.ExecSQLRequest
	clusterConfig  map[string]string
	operatorInsCli pb.MySQLOperatorInstanceServiceServer
}

type Option func(o *commonInstanceAdapter)

func WithUseOperator(useOperator bool) Option {
	return func(o *commonInstanceAdapter) {
		o.useOperator = useOperator
	}
}

func WithOperatorCli(cli pb.MySQLOperatorInstanceServiceServer) Option {
	if cli == nil {
		panic("pb.MySQLOperatorInstanceServiceServer is nil, is it not autowired ?")
	}
	return func(o *commonInstanceAdapter) {
		o.operatorInsCli = cli
	}
}

func WithUsername(username string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.Username = username
	}
}

func WithPassword(password string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.Password = password
	}
}

func WithSchema(schema string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.Schema = schema
	}
}

func WithClusterKey(clusterKey string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.ClusterKey = clusterKey
	}
}

func WithAddress(host string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.WAddress = host
	}
}

func WithQueries(queries []string) Option {
	return func(cia *commonInstanceAdapter) {
		cia.execSQLRequest.Queries = queries
	}
}
