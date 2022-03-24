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

package issue

import (
	"testing"
	"time"

	"github.com/erda-project/erda/modules/dop/dao"
)

func Test_inIterationInterval(t *testing.T) {
	type args struct {
		iteration *dao.Iteration
		t         *time.Time
	}
	t1, _ := time.Parse("2006-01-02 15:04:05", "2022-02-01 12:12:12.0")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2022-03-01 18:19:20.0")
	t3, _ := time.Parse("2006-01-02 15:04:05", "2022-03-02 18:19:20.0")
	t4, _ := time.Parse("2006-01-02 15:04:05", "2022-03-01 22:19:20.0")
	t5, _ := time.Parse("2006-01-02 15:04:05", "2022-02-10 22:19:20.0")
	iteration := &dao.Iteration{
		StartedAt:  &t1,
		FinishedAt: &t2,
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"inclusion valid",
			args{
				iteration,
				&t5,
			},
			true,
		},
		{
			"boundary valid",
			args{
				iteration,
				&t4,
			},
			true,
		},
		{
			"invalid",
			args{
				iteration,
				&t3,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inIterationInterval(tt.args.iteration, tt.args.t); got != tt.want {
				t.Errorf("inIterationInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
