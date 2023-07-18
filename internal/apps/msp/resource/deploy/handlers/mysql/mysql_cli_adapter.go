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
	var ia = instanceAdapter{
		useOperator:    false,
		execSQLRequest: new(pb.ExecSQLRequest),
		clusterConfig:  make(map[string]string),
		operatorInsCli: nil,
	}
	for _, f := range options {
		f(&ia)
	}
	if ia.useOperator {
		host, _, _ := net.SplitHostPort(ia.execSQLRequest.WAddress)
		ia.execSQLRequest.WAddress = host + ":33080" // 33080 is the mylet port // todo: hard coed here
		return &operatorClient{ia: &ia}
	}
	return &localClient{ia: &ia}
}

type localClient struct {
	ia *instanceAdapter
}

func (l *localClient) ExecSQLs() error {
	return (&mysqlhelper.Request{
		ClusterKey: l.ia.execSQLRequest.GetClusterKey(),
		Url:        "jdbc:mysql://" + l.ia.execSQLRequest.GetWAddress(),
		User:       l.ia.execSQLRequest.GetUsername(),
		Password:   l.ia.execSQLRequest.GetPassword(),
		Sqls:       l.ia.execSQLRequest.GetQueries(),
	}).Exec()
}

type operatorClient struct {
	ia *instanceAdapter
}

func (o *operatorClient) ExecSQLs() error {
	o.ia.execSQLRequest.QueryType = pb.QueryType_sql // only support pb.QueryType_sql yet
	_, err := o.ia.operatorInsCli.ExecSQL(context.Background(), o.ia.execSQLRequest)
	return err
}

type instanceAdapter struct {
	useOperator    bool
	execSQLRequest *pb.ExecSQLRequest
	clusterConfig  map[string]string
	operatorInsCli pb.MySQLOperatorInstanceServiceServer
}

type Option func(o *instanceAdapter)

func WithUseOperator(useOperator bool) Option {
	return func(o *instanceAdapter) {
		o.useOperator = useOperator
	}
}

func WithOperatorCli(cli pb.MySQLOperatorInstanceServiceServer) Option {
	if cli == nil {
		panic("pb.MySQLOperatorInstanceServiceServer is nil, is it not autowired ?")
	}
	return func(o *instanceAdapter) {
		o.operatorInsCli = cli
	}
}

func WithUsername(username string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.Username = username
	}
}

func WithPassword(password string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.Password = password
	}
}

func WithSchema(schema string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.Schema = schema
	}
}

func WithClusterKey(clusterKey string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.ClusterKey = clusterKey
	}
}

func WithAddress(host string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.WAddress = host
	}
}

func WithQueries(queries []string) Option {
	return func(ia *instanceAdapter) {
		ia.execSQLRequest.Queries = queries
	}
}
