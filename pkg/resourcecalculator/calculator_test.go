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

package resourcecalculator_test

import (
	"encoding/json"
	"testing"

	calcu "github.com/erda-project/erda/pkg/resourcecalculator"
)

type result struct {
	TotalQuotable   uint64
	ProdQuotable    uint64
	StagingQuotable uint64
	TestQuotable    uint64
	DevQuota        uint64
}

func TestResourceCalculator_AddValue(t *testing.T) {
	t.Run("Case-1", testResourceCalculator_AddValue_Case1)
	t.Run("Case-2", testResourceCalculator_AddValue_Case2)
	t.Run("Case-3", testResourceCalculator_AddValue_Case3)
	t.Run("Case-4", testResourceCalculator_AddValue_Case4)
}

func testResourceCalculator_AddValue_Case1(t *testing.T) {
	c := initC(t)
	r := result{
		TotalQuotable:   7,
		ProdQuotable:    1,
		StagingQuotable: 3,
		TestQuotable:    3,
		DevQuota:        4,
	}
	if err := c.Mem.Quota(calcu.Dev, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Test, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Staging, 1); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	t.Logf("new c: %s", string(data))
	t.Logf("quotable: total: %v, dev: %v, test: %v, staging: %v",
		c.Mem.TotalQuotable(), c.Mem.TotalForWorkspace(calcu.Dev), c.Mem.TotalForWorkspace(calcu.Test), c.Mem.TotalForWorkspace(calcu.Staging))

	if c.Mem.TotalQuotable() != r.TotalQuotable {
		t.Error("total error")
	}
	if c.Mem.TotalForWorkspace(calcu.Staging) != r.StagingQuotable {
		t.Errorf("staging error")
	}
	if c.Mem.TotalForWorkspace(calcu.Test) != r.TestQuotable {
		t.Error("test error")
	}
	if c.Mem.TotalForWorkspace(calcu.Dev) != r.DevQuota {
		t.Error("dev error")
	}
}

func testResourceCalculator_AddValue_Case2(t *testing.T) {
	c := initC(t)
	r := result{
		TotalQuotable:   5,
		ProdQuotable:    1,
		StagingQuotable: 3,
		TestQuotable:    1,
		DevQuota:        3,
	}
	if err := c.Mem.Quota(calcu.Dev, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Test, 3); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Staging, 1); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	t.Logf("new c: %s", string(data))
	t.Logf("quotable: total: %v, dev: %v, test: %v, staging: %v",
		c.Mem.TotalQuotable(), c.Mem.TotalForWorkspace(calcu.Dev), c.Mem.TotalForWorkspace(calcu.Test), c.Mem.TotalForWorkspace(calcu.Staging))

	if c.Mem.TotalQuotable() != r.TotalQuotable {
		t.Error("total error")
	}
	if c.Mem.TotalForWorkspace(calcu.Staging) != r.StagingQuotable {
		t.Error("staging error")
	}
	if c.Mem.TotalForWorkspace(calcu.Test) != r.TestQuotable {
		t.Error("test error")
	}
	if c.Mem.TotalForWorkspace(calcu.Dev) != r.DevQuota {
		t.Error("dev error")
	}
}

func testResourceCalculator_AddValue_Case3(t *testing.T) {
	c := initC(t)
	r := result{
		TotalQuotable:   4,
		ProdQuotable:    1,
		StagingQuotable: 2,
		TestQuotable:    0,
		DevQuota:        2,
	}
	if err := c.Mem.Quota(calcu.Dev, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Test, 4); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Staging, 1); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	t.Logf("new c: %s", string(data))
	t.Logf("quotable: total: %v, dev: %v, test: %v, staging: %v",
		c.Mem.TotalQuotable(), c.Mem.TotalForWorkspace(calcu.Dev), c.Mem.TotalForWorkspace(calcu.Test), c.Mem.TotalForWorkspace(calcu.Staging))

	if c.Mem.TotalQuotable() != r.TotalQuotable {
		t.Error("total error")
	}
	if c.Mem.TotalForWorkspace(calcu.Staging) != r.StagingQuotable {
		t.Error("staging error")
	}
	if c.Mem.TotalForWorkspace(calcu.Test) != r.TestQuotable {
		t.Error("test error")
	}
	if c.Mem.TotalForWorkspace(calcu.Dev) != r.DevQuota {
		t.Error("dev error")
	}
}

func testResourceCalculator_AddValue_Case4(t *testing.T) {
	c := initC(t)
	r := result{
		TotalQuotable:   3,
		ProdQuotable:    1,
		StagingQuotable: 2,
		TestQuotable:    0,
		DevQuota:        1,
	}
	if err := c.Mem.Quota(calcu.Dev, 3); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Test, 3); err != nil {
		t.Fatal(err)
	}
	if err := c.Mem.Quota(calcu.Staging, 1); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	t.Logf("new c: %s", string(data))
	t.Logf("quotable: total: %v, dev: %v, test: %v, staging: %v",
		c.Mem.TotalQuotable(), c.Mem.TotalForWorkspace(calcu.Dev), c.Mem.TotalForWorkspace(calcu.Test), c.Mem.TotalForWorkspace(calcu.Staging))

	if c.Mem.TotalQuotable() != r.TotalQuotable {
		t.Error("total error")
	}
	if c.Mem.TotalForWorkspace(calcu.Staging) != r.StagingQuotable {
		t.Error("staging error")
	}
	if c.Mem.TotalForWorkspace(calcu.Test) != r.TestQuotable {
		t.Error("test error")
	}
	if c.Mem.TotalForWorkspace(calcu.Dev) != r.DevQuota {
		t.Error("dev error")
	}
}

func initC(t *testing.T) *calcu.Calculator {
	clusterName := "erda-hongkong"
	c := calcu.New(clusterName)
	c.Mem.AddValue(2, calcu.Dev)
	c.Mem.AddValue(1, calcu.Dev, calcu.Test)
	c.Mem.AddValue(1, calcu.Dev, calcu.Prod)
	c.Mem.AddValue(1, calcu.Dev, calcu.Test, calcu.Staging)
	c.Mem.AddValue(2, calcu.Test)
	c.Mem.AddValue(3, calcu.Staging)
	data, _ := json.MarshalIndent(c.Mem, "", "  ")
	t.Logf("initC total quotable: %v, c: %s", c.Mem.TotalQuotable(), string(data))
	return c
}

func TestWorkspaceString(t *testing.T) {
	for _, workspace := range calcu.Workspaces {
		t.Log(calcu.WorkspaceString(workspace))
	}
	if calcu.WorkspaceString(4) != "" {
		t.Log("default value error")
	}
}

func TestCoreToMillcore(t *testing.T) {
	var (
		v float64 = 1.2
		r uint64  = 1200
	)
	if calcu.CoreToMillcore(v) != r {
		t.Fatal("calculate error")
	}
}

func TestMillcoreToCore(t *testing.T) {
	var (
		v uint64  = 1200
		r float64 = 1.2
	)
	if calcu.MillcoreToCore(v) != r {
		t.Fatal("calculate error")
	}
}

func TestGibibyteToByte(t *testing.T) {
	var (
		v float64 = 1.5
		r uint64  = 1.5 * 1024 * 1024 * 1024
	)
	if calcu.GibibyteToByte(v) != r {
		t.Fatal("calculate error")
	}
}

func TestByteToGibibyte(t *testing.T) {
	var (
		v uint64  = 1.5 * 1024 * 1024 * 1024
		r float64 = 1.5
	)
	if calcu.ByteToGibibyte(v) != r {
		t.Fatal("calculate error")
	}
}

func TestWorkspacesString(t *testing.T) {
	t.Log(calcu.WorkspacesString([]calcu.Workspace{calcu.Prod, calcu.Dev, calcu.Staging}))
}

func TestResourceCalculator_Copy(t *testing.T) {
	clusterName := "erda-hongkong"
	c := calcu.New(clusterName)
	d := c.Copy()
	for k, v := range c.CPU.M {
		if v != d.CPU.M[k] {
			t.Fatal("copy error")
		}
	}
}

func TestResourceCalculator_AlreadyQuota(t *testing.T) {
	clusterName := "erda-hongkong"
	c := calcu.New(clusterName)
	if c.CPU.AlreadyQuota(calcu.Prod) != 0 {
		t.Fatal("AlreadyQuota error")
	}
}

func TestResourceCalculator_StatusOK(t *testing.T) {
	clusterName := "erda-hongkong"
	c := calcu.New(clusterName)
	if !c.CPU.StatusOK(calcu.Prod) {
		t.Fatal("StatusOK error")
	}
}

func TestResourceToString(t *testing.T) {
	t.Log(calcu.ResourceToString(1000, "cpu"))
	t.Log(calcu.ResourceToString(5*1024*1024*1024, "memory"))
	t.Log(calcu.ResourceToString(1000, "error key"))
}
