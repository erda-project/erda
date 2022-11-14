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

package canal

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/sourcecov/mock"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type k8s struct{}

func (k8s) GetK8SAddr() string {
	return ""
}

func TestCanalOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ns := mock.NewMockNamespaceUtil(ctrl)

	mo := New(new(k8s), ns, nil, nil, httpclient.New())
	sg := new(apistructs.ServiceGroup)
	sg.Services = append(sg.Services, apistructs.Service{
		Name: "canal",
	})
	sg.ID = "abcdefghigklmn"
	mo.Name(sg)
	mo.NamespacedName(sg)
	mo.IsSupported()
	mo.Validate(sg)
	sg.Labels = make(map[string]string)
	sg.Labels["USE_OPERATOR"] = "canal"
	mo.Validate(sg)
	sg.Services[0].Env = make(map[string]string)
	mo.Validate(sg)
	sg.Services[0].Env["CANAL_DESTINATION"] = "b"
	sg.Services[0].Env["canal.instance.master.address"] = "1"
	sg.Services[0].Env["canal.instance.master.address"] = "1"
	sg.Services[0].Env["canal.instance.dbUsername"] = "2"
	sg.Services[0].Env["canal.instance.dbPassword"] = "3"
	mo.Validate(sg)
	mo.Convert(sg)
	mo.Create(sg)
	mo.Inspect(sg)
	mo.Update(sg)
	mo.Remove(sg)
}
