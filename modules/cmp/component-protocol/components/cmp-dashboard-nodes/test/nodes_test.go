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

package test

import (
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
)

var scenario apistructs.ComponentProtocol
var ctxBdl *bundle.Bundle

func init() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: "conf/cmp/cmp.yaml",
	})
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f, err := os.Open(pwd + "/protocol.yml")
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		panic(err)
	}
	fmt.Printf("scenario key = %s\n", scenario.Scenario)

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
	ctxBdl = *bundle.Bundle{
		Bdl:         bdl,
		I18nPrinter: nil,
		InParams:    req.SDK.InParams,
	}
}
