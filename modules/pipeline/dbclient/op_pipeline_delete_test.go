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

package dbclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_canDelete(t *testing.T) {
	type args struct {
		p spec.Pipeline
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "test_can_not_delete",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status: apistructs.PipelineStatusRunning,
					},
				},
			},
			want:  false,
			want1: fmt.Sprintf("invalid status: %s", apistructs.PipelineStatusRunning),
		},
		{
			name: "test_can_not_analyzed",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status: apistructs.PipelineStatusAnalyzed,
					},
				},
			},
			want:  true,
			want1: "",
		},

		{
			name: "test_end_status_not_CompleteReconcilerGC",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status: apistructs.PipelineStatusFailed,
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CompleteReconcilerGC: false,
						},
					},
				},
			},
			want:  false,
			want1: fmt.Sprintf("waiting gc"),
		},
		{
			name: "test_end_status_CompleteReconcilerGC",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status: apistructs.PipelineStatusFailed,
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CompleteReconcilerGC: true,
						},
					},
				},
			},
			want:  true,
			want1: "",
		},
		{
			name: "test_end_status_not_CompleteReconcilerGC_but_end_gc_time",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status:  apistructs.PipelineStatusFailed,
						TimeEnd: &[]time.Time{time.Now().Add(-5184555 * time.Second)}[0],
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							GC: apistructs.PipelineGC{
								DatabaseGC: apistructs.PipelineDatabaseGC{
									Finished: apistructs.PipelineDBGCItem{
										TTLSecond: &[]uint64{5184000}[0],
									},
								},
							},
							CompleteReconcilerGC: false,
						},
					},
				},
			},
			want:  true,
			want1: "",
		},
		{
			name: "test_end_status_not_CompleteReconcilerGC_but_not_end_gc_time",
			args: args{
				p: spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						Status:  apistructs.PipelineStatusFailed,
						TimeEnd: &[]time.Time{time.Now().Add(-5182555 * time.Second)}[0],
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							GC: apistructs.PipelineGC{
								DatabaseGC: apistructs.PipelineDatabaseGC{
									Finished: apistructs.PipelineDBGCItem{
										TTLSecond: &[]uint64{5184000}[0],
									},
								},
							},
							CompleteReconcilerGC: false,
						},
					},
				},
			},
			want:  false,
			want1: fmt.Sprintf("waiting gc"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := canDelete(tt.args.p)
			if got != tt.want {
				t.Errorf("canDelete() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("canDelete() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
