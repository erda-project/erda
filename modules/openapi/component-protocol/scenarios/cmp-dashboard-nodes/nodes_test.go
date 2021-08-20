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

package cmp_dashboard_nodes

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"io"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/components/memTable"
)

var scenario apistructs.ComponentProtocol
var ctxBdl protocol.ContextBundle

func init(){
	common.Run(&servicehub.RunOptions{})
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f ,err := os.Open(pwd+"/protocol.yml")
	if err!= nil{
		panic(err)
	}
	data ,err := io.ReadAll(f)
	if err != nil{
		panic(err)
	}
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		panic(err)
	}
	fmt.Printf("scenario key = %s\n",scenario.Scenario)

	req := apistructs.ComponentProtocolRequest{}

	var bundleOpts = []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*90),
				httpclient.WithEnableAutoRetry(false),
			)),
		bundle.WithAllAvailableClients(),
	}
	bdl := bundle.New(bundleOpts...)
	// get locale from request
	ctxBdl = protocol.ContextBundle{
		Bdl:         bdl,
		I18nPrinter: nil,
		InParams:    req.InParams,
	}
}

func TestTable(t *testing.T){

	memTable := memTable.RenderCreator()
	ce := apistructs.ComponentEvent{}
	gs := apistructs.GlobalStateData{}
	c := scenario.Components["memTable"]
	memTable.Render(context.WithValue(context.Background(),protocol.GlobalInnerKeyCtxBundle.String(),ctxBdl), c , apistructs.ComponentProtocolScenario{}, ce, &gs)
}
