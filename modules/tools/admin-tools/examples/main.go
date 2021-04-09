// Copyright (c) 2021 Terminus, Inc.
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

package main

import (
	"fmt"
	"os"

	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/elasticsearch"
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda/modules/tools/admin-tools"
	"github.com/olivere/elastic"
)

type define struct{}

type provider struct {
	es *elastic.Client
}

func (d *define) Services() []string     { return []string{"hello"} }
func (d *define) Dependencies() []string { return []string{"admin-tools"} }
func (d *define) Description() string    { return "hello for example" }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

func (p *provider) Init(ctx servicehub.Context) error {
	fmt.Println(p.es)
	return nil
}

func init() {
	servicehub.RegisterProvider("example", &define{})
}

func main() {
	hub := servicehub.New()
	hub.Run("examples", "", os.Args...)
}
