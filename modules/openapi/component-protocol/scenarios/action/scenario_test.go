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

package action

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	_ "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/action/components/actionForm"
	"github.com/erda-project/erda/modules/openapi/i18n"
	"github.com/erda-project/erda/pkg/httpclient"
)

func rend(req *apistructs.ComponentProtocolRequest) (cont *apistructs.ComponentProtocolRequest, err error) {
	// bundle
	os.Setenv("DICEHUB_ADDR", "dicehub.default.svc.cluster.local:10000")
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithDiceHub(),
	}
	bdl := bundle.New(bundleOpts...)
	r := http.Request{}
	i18nPrinter := i18n.I18nPrinter(&r)
	ctxBdl := protocol.ContextBundle{
		Bdl:         bdl,
		I18nPrinter: i18nPrinter,
		InParams:    req.InParams,
	}
	ctx := context.Background()
	ctx1 := context.WithValue(ctx, protocol.GlobalInnerKeyCtxBundle.String(), ctxBdl)

	err = protocol.RunScenarioRender(ctx1, req)
	if err != nil {
		return
	}
	cont = req
	return
}

func TestStateInjectLess(t *testing.T) {
	str1 := "{\"key1\":\"value1\",\"key2\":\"value2\"}"
	type s1 struct {
		Key1 string `json:"key1"`
	}
	type s2 struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}
	type s3 struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
		Key3 string `json:"key3"`
	}

	t1 := s1{}
	t2 := s2{}
	t3 := s3{}

	err := json.Unmarshal([]byte(str1), &t1)
	if err != nil {
		t.Logf("unmarshal str1 to t1 failed, err:%v", err)
	}
	err = json.Unmarshal([]byte(str1), &t2)
	if err != nil {
		t.Logf("unmarshal str1 to t2 failed, err:%v", err)
	}
	err = json.Unmarshal([]byte(str1), &t3)
	if err != nil {
		t.Logf("unmarshal str1 to t3 failed, err:%v", err)
	}
}
