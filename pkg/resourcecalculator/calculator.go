//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package resourcecalculator

import (
	"sort"

	"github.com/pkg/errors"
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
	ClusterName string
	CPU         *ResourceCalculator
	Mem         *ResourceCalculator
}

// Todo: 该算法并没有体现调度的优先级，需优化

type ResourceCalculator struct {
	q [4][4][4][4]uint64
}

func New(clusterName string) *Calculator {
	return &Calculator{
		ClusterName: clusterName,
		CPU:         newResourceQuota(),
		Mem:         newResourceQuota(),
	}
}

func newResourceQuota() *ResourceCalculator {
	return &ResourceCalculator{
		q: [4][4][4][4]uint64{},
	}
}

func (q *ResourceCalculator) AddValue(value uint64, workspace ...Workspace) {
	sort.Slice(workspace, func(i, j int) bool {
		return workspace[i] < workspace[j]
	})
	switch len(workspace) {
	case 0:
	case 1:
		q.q[workspace[0]][workspace[0]][workspace[0]][workspace[0]] += value
	case 2:
		q.q[workspace[0]][workspace[1]][workspace[0]][workspace[0]] += value
	case 3:
		q.q[workspace[0]][workspace[1]][workspace[2]][workspace[0]] += value
	default:
		q.q[workspace[0]][workspace[1]][workspace[2]][workspace[3]] += value
	}
}

func (q ResourceCalculator) TotalQuotable() uint64 {
	var sum uint64
	for i := 0; i < len(q.q); i++ {
		for j := 0; j < len(q.q[i]); j++ {
			for k := 0; k < len(q.q[i][j]); k++ {
				for l := 0; l < len(q.q[i][j][k]); l++ {
					sum += q.q[i][j][k][l]
				}
			}
		}
	}
	return sum
}

func (q *ResourceCalculator) TotalForWorkspace(workspace Workspace) uint64 {
	w := int(workspace)
	exclusive := q.q[w][w][w][w]
	var sum uint64

	for i := 0; i < len(q.q); i++ {
		for j := 0; j < len(q.q[i]); j++ {
			for k := 0; k < len(q.q[i][j]); k++ {
				for l := 0; l < len(q.q[i][j][k]); l++ {
					if (i == w || j == w || k == w || l == w) && !(i == w && j == w && k == w && l == w) {
						sum += q.q[i][j][k][l]
					}
				}
			}
		}
	}
	return sum + exclusive
}

func (q *ResourceCalculator) Quota(workspace Workspace, quota uint64) error {
	w := int(workspace)
	if quota > q.TotalForWorkspace(workspace) {
		q.q = [4][4][4][4]uint64{}
		return errors.New("总资源不够")
	}

	// 如果独占的已经够了
	exclusive := q.q[w][w][w][w]
	if exclusive >= quota {
		q.q[w][w][w][w] -= quota
		return nil
	}

	// 如果独占的不够, 先扣除独占部分, 再寻求公用部分
	quota -= exclusive
	q.q[w][w][w][w] = 0
	for i := 0; i < len(q.q); i++ {
		for j := 0; j < len(q.q[i]); j++ {
			for k := 0; k < len(q.q[i][j]); k++ {
				for l := 0; l < len(q.q[i][j][k]); l++ {
					if i == w || j == w || k == w || l == w {
						quotable := q.q[i][j][k][l]
						if quotable >= quota {
							q.q[i][j][k][l] -= quota
							return nil
						}
						q.q[i][j][k][l] = 0
						quota -= quotable
					}
				}
			}
		}
	}

	return nil
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

func CoreToMillcore(v float64) uint64 {
	return uint64(v * 1000)
}

func MillcoreToCore(v uint64) float64 {
	return float64(v) / 1000
}

func GibibyteToByte(v float64) uint64 {
	return uint64(v*1024*1024*1024)
}
func ByteToGibibyte(v uint64) float64 {
	return float64(v) / (1024*1024*1024)
}