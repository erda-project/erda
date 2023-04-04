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
	"testing"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
)

func TestNewInstanceAdapter(t *testing.T) {
	var configs = []Option{
		WithUsername("dspo"),
		WithPassword("password"),
		WithSchema("erda"),
		WithClusterKey("clusterKey"),
		WithAddress("mysql.local"),
		WithQueries([]string{"show tables;"}),
	}
	var _ = NewInstanceAdapter(append(configs, WithUseOperator(false))...)
	var _ = NewInstanceAdapter(append(configs, WithUseOperator(true), WithOperatorCli(new(mockOperatorCli)))...)
	// test success if not panic
}

type mockOperatorCli struct {
}

func (m mockOperatorCli) ExecSQL(context.Context, *pb.ExecSQLRequest) (*pb.ExecSQLResponse, error) {
	return nil, errors.New("this is mocked")
}
