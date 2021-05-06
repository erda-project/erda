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
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
)

//func TestBuildPack(t *testing.T) {
//	os.Setenv("DICEHUB_ADDR", "dicehub.default.svc.cluster.local:10000")
//	bundleOpts := []bundle.Option{
//		bundle.WithHTTPClient(
//			httpclient.New(
//				httpclient.WithTimeout(time.Second, time.Second*60),
//			)),
//		bundle.WithDiceHub(),
//	}
//	bdl := bundle.New(bundleOpts...)
//	comp := ComponentAction{
//		ctxBdl: protocol.ContextBundle{
//			Bdl:         bdl,
//			I18nPrinter: nil,
//			Identity:    apistructs.Identity{},
//			InParams:    nil,
//		},
//	}
//	err := comp.GenActionProps("buildpack", "")
//	if err != nil {
//		t.Errorf("generate rops failed, error:%v", err)
//	}
//	t.Logf("param props: %+v", comp.Props)
//	cont, _ := json.Marshal(comp.Props)
//	t.Logf("content:%s", string(cont))
//}

func TestLoadProtocolFromFile(t *testing.T) {
	path := "../../protocol.yml"
	f, err := os.Open(path)
	if err != nil {
		t.Errorf("open file failed, err:%v", err)
		return
	}
	proto := apistructs.ComponentProtocol{}
	err = yaml.NewDecoder(f).Decode(&proto)
	if err != nil {
		t.Errorf("decode protocol failed, err:%v", err)
		return
	}
}

var protoStr = `
hierarchy:
  root: actionForm

components:
  actionForm:
    type: "Form"
    props: "【后端动态注入]"
    data: {}
    operations:
      change:
        reload: true
    state:
      version: "[前端选择按钮输入]"
`

func TestLoadProtocolFromString(t *testing.T) {
	proto := apistructs.ComponentProtocol{}
	// bt, _ := yaml.Marshal(protoStr)
	err := yaml.Unmarshal([]byte(protoStr), &proto)
	if err != nil {
		t.Errorf("decode protocol failed, err:%v", err)
		return
	}
}
