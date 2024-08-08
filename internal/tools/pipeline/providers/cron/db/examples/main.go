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

package main

import (
	"context"
	"fmt"
	"os"

	"xorm.io/xorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
)

type provider struct {
	DB    *xorm.Engine        // autowired
	MySQL mysqlxorm.Interface // autowired

	cronDBClient *Client
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.cronDBClient = &Client{Interface: p.MySQL}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	item := PipelineCron{}
	// create
	err := p.cronDBClient.CreatePipelineCron(&item)
	if err != nil {
		return err
	}
	fmt.Println("item id: ", item.ID)

	// get
	_, found, err := p.cronDBClient.GetPipelineCron(item.ID)
	if err != nil {
		return err
	}
	fmt.Println("found: ", found)

	//// delete
	//err = p.cronDBClient.DeletePipelineCron(item.ID)
	//if err != nil {
	//	return err
	//}

	// batch delete
	err = p.cronDBClient.BatchDeletePipelineCron([]uint64{item.ID})
	if err != nil {
		return err
	}

	// get
	_, found, err = p.cronDBClient.GetPipelineCron(item.ID)
	if err != nil {
		return err
	}
	fmt.Println("found after delete: ", found)

	return nil
}

func init() {
	servicehub.Register("example", &servicehub.Spec{
		Services:     []string{"example"},
		Dependencies: []string{"mysql-xorm"},
		Description:  "example",
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}

func main() {
	hub := servicehub.New()
	hub.Run("examples", "", os.Args...)
}
