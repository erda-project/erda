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

package apistructs_test

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestGetQuotaOnClustersResponse_ReCalcu(t *testing.T) {
	var project = apistructs.ProjectQuotaOnClusters{
		ID:          0,
		Name:        "project-1",
		DisplayName: "project-2",
		CPUQuota:    0,
		MemQuota:    0,
	}
	var owner = apistructs.OwnerQuotaOnClusters{
		ID:       0,
		Name:     "erda",
		Nickname: "erda",
		CPUQuota: 0,
		MemQuota: 0,
		Projects: nil,
	}
	owner.Projects = []*apistructs.ProjectQuotaOnClusters{&project}
	var resp apistructs.GetQuotaOnClustersResponse
	resp.ClusterNames = []string{"erda-hongkong"}
	resp.Owners = []*apistructs.OwnerQuotaOnClusters{&owner}

	project.AccuQuota(1000, 1024*1024*1024)
	resp.ReCalcu()
	t.Logf("resp: %+v", resp)
	t.Logf("project: %+v", project)
	if resp.CPUQuota != 1 {
		t.Fatal("cpu quota error")
	}
	if resp.MemQuota != 1 {
		t.Fatal("mem quota error")
	}
}
