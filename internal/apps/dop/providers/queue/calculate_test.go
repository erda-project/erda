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

package queue

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func Test_calculateProjectResource(t *testing.T) {
	type args struct {
		workspace string
		project   *apistructs.ProjectDTO
		appCount  int64
	}
	tests := []struct {
		name string
		args args
		want *ProjectQueueResource
	}{
		{
			name: "test calculate project resource",
			args: args{
				workspace: "DEV",
				project: &apistructs.ProjectDTO{
					ID: 1,
					ResourceConfig: &apistructs.ResourceConfigsInfo{
						DEV: &apistructs.ResourceConfigInfo{
							CPUQuota: 30,
							MemQuota: 30,
						},
					},
				},
				appCount: 30,
			},
			want: &ProjectQueueResource{
				Concurrency: 60,
				MaxCPU:      30,
				MaxMemoryMB: 30720,
			},
		},
		{
			name: "cpu lower than default",
			args: args{
				workspace: "DEV",
				project: &apistructs.ProjectDTO{
					ID: 1,
					ResourceConfig: &apistructs.ResourceConfigsInfo{
						DEV: &apistructs.ResourceConfigInfo{
							CPUQuota: 10,
							MemQuota: 30,
						},
					},
				},
				appCount: 30,
			},
			want: &ProjectQueueResource{
				Concurrency: 60,
				MaxCPU:      20,
				MaxMemoryMB: 30720,
			},
		},
		{
			name: "concurrency lower than default",
			args: args{
				workspace: "DEV",
				project: &apistructs.ProjectDTO{
					ID: 1,
					ResourceConfig: &apistructs.ResourceConfigsInfo{
						DEV: &apistructs.ResourceConfigInfo{
							CPUQuota: 30,
							MemQuota: 30,
						},
					},
				},
				appCount: 5,
			},
			want: &ProjectQueueResource{
				Concurrency: 20,
				MaxCPU:      30,
				MaxMemoryMB: 30720,
			},
		},
	}
	p := &provider{
		Cfg: &config{
			DefaultQueueConcurrency: 20,
			DefaultQueueCPU:         20,
			DefaultQueueMemoryMB:    20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdl := bundle.New()
			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CountAppByProID", func(_ *bundle.Bundle, _ uint64) (int64, error) {
				return tt.args.appCount, nil
			})
			p.bdl = bdl
			got, err := p.calculateProjectResource(tt.args.workspace, tt.args.project)
			if err != nil {
				t.Errorf("calculateProjectResource() error = %v", err)
			}
			if got.MaxCPU != tt.want.MaxCPU {
				t.Errorf("calculateProjectResource() cpu resource got = %v, want %v", got.MaxCPU, tt.want.MaxCPU)
			}
			if got.Concurrency != tt.want.Concurrency {
				t.Errorf("calculateProjectResource() concurrency resource got = %v, want %v", got.Concurrency, tt.want.Concurrency)
			}
			if got.MaxMemoryMB != tt.want.MaxMemoryMB {
				t.Errorf("calculateProjectResource() memory MB resource got = %v, want %v", got.MaxMemoryMB, tt.want.MaxMemoryMB)
			}
		})
	}
}
