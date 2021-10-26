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

package resourcecalculator

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	Dev Workspace = iota
	Test
	Staging
	Prod
)

var Workspaces = []Workspace{Prod, Staging, Test, Dev}

type Workspace int

type Calculator struct {
	ClusterName    string
	allocatableCPU *ResourceCalculator
	allocatableMem *ResourceCalculator
	availableCPU   *ResourceCalculator
	availableMem   *ResourceCalculator
}

func New(clusterName string) *Calculator {
	return &Calculator{
		ClusterName: clusterName,
		allocatableCPU: &ResourceCalculator{
			Type:  "CPU",
			M:     make(map[string]uint64),
			quota: make(map[Workspace]uint64),
		},
		availableCPU: &ResourceCalculator{
			Type:  "CPU",
			M:     make(map[string]uint64),
			quota: make(map[Workspace]uint64),
		},
		allocatableMem: &ResourceCalculator{
			Type:  "Memory",
			M:     make(map[string]uint64),
			quota: make(map[Workspace]uint64),
		},
		availableMem: &ResourceCalculator{
			Type:  "Memory",
			M:     make(map[string]uint64),
			quota: make(map[Workspace]uint64),
		},
	}
}

func (c *Calculator) AddValue(cpu, mem uint64, workspace ...Workspace) {
	c.allocatableCPU.addValue(cpu, workspace...)
	c.availableCPU.addValue(cpu, workspace...)
	c.allocatableMem.addValue(mem, workspace...)
	c.availableMem.addValue(mem, workspace...)
}

func (c *Calculator) DeductionQuota(workspace Workspace, cpu, mem uint64) {
	c.availableCPU.deductionQuota(workspace, cpu)
	c.availableMem.deductionQuota(workspace, mem)
}

func (c *Calculator) AllocatableCPU(workspace Workspace) uint64 {
	return c.allocatableCPU.totalForWorkspace(workspace)
}

func (c *Calculator) AllocatableMem(workspace Workspace) uint64 {
	return c.allocatableMem.totalForWorkspace(workspace)
}

func (c *Calculator) AlreadyQuotaCPU(workspace Workspace) uint64 {
	return c.availableCPU.alreadyQuota(workspace)
}

func (c *Calculator) AlreadyQuotaMem(workspace Workspace) uint64 {
	return c.availableMem.alreadyQuota(workspace)
}

func (c *Calculator) TotalQuotableCPU() uint64 {
	var total = int(c.allocatableCPU.total)
	for _, v := range c.availableCPU.quota {
		total -= int(v)
	}
	if total < 0 {
		total = 0
	}
	return uint64(total)
}

func (c *Calculator) TotalQuotableMem() uint64 {
	var total = int(c.allocatableMem.total)
	for _, v := range c.availableMem.M {
		total -= int(v)
	}
	if total < 0 {
		total = 0
	}
	return uint64(total)
}

func (c *Calculator) QuotableCPUForWorkspace(workspace Workspace) uint64 {
	return c.availableCPU.totalForWorkspace(workspace)
}

func (c *Calculator) QuotableMemForWorkspace(workspace Workspace) uint64 {
	return c.availableMem.totalForWorkspace(workspace)
}

type ResourceCalculator struct {
	Type  string
	M     map[string]uint64
	quota map[Workspace]uint64
	total uint64
}

func (q *ResourceCalculator) addValue(value uint64, workspace ...Workspace) {
	q.total += value
	workspaces := WorkspacesString(workspace)
	if length := len(workspaces); length == 0 || length > 4 {
		return
	}
	w := strings.Join(workspaces, ":")
	q.M[w] += value
}

func (q *ResourceCalculator) totalForWorkspace(workspace Workspace) uint64 {
	var (
		sum uint64
		w   = WorkspaceString(workspace)
	)
	if w == "" {
		return 0
	}
	for k, v := range q.M {
		if strings.Contains(k, w) {
			sum += v
		}
	}
	return sum
}

func (q *ResourceCalculator) deductionQuota(workspace Workspace, quota uint64) {
	q.quota[workspace] += quota
	if totalForWorkspace := q.totalForWorkspace(workspace); quota > totalForWorkspace {
		for k := range q.M {
			if strings.Contains(k, WorkspaceString(workspace)) {
				q.M[k] = 0
			}
		}
		return
	}

	// 按优先级减扣
	p := priority(workspace)
	for _, v := range p {
		if q.M[v] >= quota {
			q.M[v] -= quota
			return
		}
		quota -= q.M[v]
		q.M[v] = 0
	}
}

func (q *ResourceCalculator) alreadyQuota(workspace Workspace) uint64 {
	return q.quota[workspace]
}

func WorkspaceString(workspace Workspace) string {
	switch workspace {
	case Prod:
		return "prod"
	case Staging:
		return "staging"
	case Test:
		return "test"
	case Dev:
		return "dev"
	default:
		return ""
	}
}

func WorkspacesString(workspaces []Workspace) []string {
	var m = make(map[Workspace]bool)
	for _, w := range workspaces {
		m[w] = true
	}
	workspaces = []Workspace{}
	for v := range m {
		workspaces = append(workspaces, v)
	}
	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i] < workspaces[j]
	})
	var result []string
	for _, v := range workspaces {
		result = append(result, WorkspaceString(v))
	}
	return result
}

func CoreToMillcore(v float64) uint64 {
	return uint64(v * 1000)
}

func MillcoreToCore(v uint64, accuracy int32) float64 {
	value, _ := decimal.NewFromFloat(float64(v) / 1000).Round(accuracy).Float64()
	return value
}

func GibibyteToByte(v float64) uint64 {
	return uint64(v * 1024 * 1024 * 1024)
}
func ByteToGibibyte(v uint64, accuracy int32) float64 {
	value, _ := decimal.NewFromFloat(float64(v) / (1024 * 1024 * 1024)).Round(accuracy).Float64()
	return value
}

func accuracy() {

}

func priority(workspace Workspace) []string {
	switch workspace {
	case Prod:
		return []string{
			"prod",
			"dev:prod", "test:prod", "staging:prod",
			"dev:test:prod", "dev:staging:prod", "test:staging:prod",
			"dev:test:staging:prod",
		}
	case Staging:
		return []string{
			"staging",
			"dev:staging", "test:staging", "staging:prod",
			"dev:test:staging", "dev:staging:prod", "test:staging:prod",
			"dev:test:staging:prod",
		}
	case Test:
		return []string{
			"test",
			"dev:test", "test:staging", "test:prod",
			"dev:test:staging", "dev:test:prod", "test:staging:prod",
			"dev:test:staging:prod",
		}
	case Dev:
		return []string{
			"dev",
			"dev:test", "dev:staging", "dev:prod",
			"dev:test:staging", "dev:test:prod", "dev:staging:prod",
			"dev:test:staging:prod",
		}
	}
	return nil
}

func ResourceToString(res float64, typ string) string {
	switch typ {
	case "cpu":
		return strconv.FormatFloat(setPrec(res/1000, 3), 'f', -1, 64)
	case "memory":
		units := []string{"B", "KB", "MB", "GB", "TB"}
		i := 0
		for res >= 1<<10 && i < len(units)-1 {
			res /= 1 << 10
			i++
		}
		return fmt.Sprintf("%s%s", strconv.FormatFloat(setPrec(res, 3), 'f', -1, 64), units[i])
	default:
		return fmt.Sprintf("%.f", res)
	}
}

func setPrec(f float64, prec int) float64 {
	pow := math.Pow10(prec)
	f = float64(int64(f*pow)) / pow
	return f
}
