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

package workbench

import (
	"os"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func initBundle() *bundle.Bundle {
	os.Setenv("CORE_SERVICES_ADDR", "http://core-services.project-387-dev.svc.cluster.local:9526")
	os.Setenv("MSP_ADDR", "http://msp.project-387-dev.svc.cluster.local:8080")
	os.Setenv("DOP_ADDR", "http://dop.project-387-dev.svc.cluster.local:9527")
	os.Setenv("ORCHESTRATOR_ADDR", "http://orchestrator.project-387-dev.svc.cluster.local:8081")
	os.Setenv("GITTAR_ADDR", "http://gittar.project-387-dev.svc.cluster.local:5566")
	bdl := bundle.New(
		bundle.WithCoreServices(),
		bundle.WithMSP(),
		bundle.WithDOP(),
		bundle.WithGittar(),
		bundle.WithOrchestrator(),
	)
	return bdl
}

func TestListProj(t *testing.T) {
	bdl := initBundle()
	wb := New(WithBundle(bdl))
	identity := apistructs.Identity{
		UserID: "2",
		OrgID:  "1",
	}

	data, err := wb.ListQueryProjWbData(identity, apistructs.PageRequest{PageNo: 1, PageSize: 10}, "")
	if err != nil {
		t.Logf("list query proj wb data faield, error: %v", err)
	}
	t.Logf("data: %v", data)
}

func TestListSub(t *testing.T) {
	bdl := initBundle()
	wb := New(WithBundle(bdl))
	identity := apistructs.Identity{
		UserID: "2",
		OrgID:  "1",
	}

	data, err := wb.ListSubProjWbData(identity)
	if err != nil {
		t.Logf("list query proj wb data faield, error: %v", err)
	}
	t.Logf("data: %+v", data)
}

func TestListApp(t *testing.T) {
	bdl := initBundle()
	wb := New(WithBundle(bdl))
	identity := apistructs.Identity{
		UserID: "2",
		OrgID:  "1",
	}

	data, err := wb.ListAppWbData(identity, apistructs.ApplicationListRequest{PageNo: 1, PageSize: 10}, 1)
	if err != nil {
		t.Logf("list query proj wb data faield, error: %v", err)
	}
	t.Logf("data: %v", data)
}

func TestGetAppMr(t *testing.T) {
	bdl := initBundle()
	// wb := New(WithBundle(bdl))
	identity := apistructs.Identity{
		UserID: "2",
		OrgID:  "1",
	}
	rsp, err := bdl.ListMergeRequest(45, identity.UserID, apistructs.GittarQueryMrRequest{})
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("response: %+v", rsp)
}

func TestStateIds(t *testing.T) {
	bdl := initBundle()
	wb := New(WithBundle(bdl))
	ids, err := wb.GetAllIssueStateIDs(3)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("ids: %v", ids)
}

func TestGetUrlQueries(t *testing.T) {
	bdl := initBundle()
	wb := New(WithBundle(bdl))
	r, err := wb.GetIssueQueries(3)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("result: %+v", r)
}
