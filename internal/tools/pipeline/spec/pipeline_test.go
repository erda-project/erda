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

package spec

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
)

func TestPipeline_EnsureGC(t *testing.T) {
	var ttl uint64 = 10
	var archive bool = true
	pipelines := []Pipeline{
		{
			PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					GC: basepb.PipelineGC{
						ResourceGC: &basepb.PipelineResourceGC{
							SuccessTTLSecond: &ttl,
							FailedTTLSecond:  nil,
						},
						DatabaseGC: &basepb.PipelineDatabaseGC{
							Analyzed: &basepb.PipelineDBGCItem{
								NeedArchive: nil,
								TTLSecond:   &ttl,
							},
							Finished: &basepb.PipelineDBGCItem{
								NeedArchive: &archive,
								TTLSecond:   nil,
							},
						},
					},
				},
			},
		},
		{
			PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					GC: basepb.PipelineGC{
						ResourceGC: nil,
						DatabaseGC: &basepb.PipelineDatabaseGC{
							Analyzed: &basepb.PipelineDBGCItem{
								NeedArchive: nil,
								TTLSecond:   &ttl,
							},
							Finished: &basepb.PipelineDBGCItem{
								NeedArchive: &archive,
								TTLSecond:   nil,
							},
						},
					},
				},
			},
		},
		{
			PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					GC: basepb.PipelineGC{
						ResourceGC: &basepb.PipelineResourceGC{
							SuccessTTLSecond: &ttl,
							FailedTTLSecond:  nil,
						},
						DatabaseGC: &basepb.PipelineDatabaseGC{
							Analyzed: &basepb.PipelineDBGCItem{
								NeedArchive: nil,
								TTLSecond:   &ttl,
							},
							Finished: &basepb.PipelineDBGCItem{
								NeedArchive: &archive,
								TTLSecond:   nil,
							},
						},
					},
				},
			},
		},
	}
	for _, pipeline := range pipelines {
		pipeline.EnsureGC()
	}
}

func TestCanSkipRunningCheck(t *testing.T) {
	tests := []struct {
		p    *Pipeline
		want bool
	}{
		{
			p: &Pipeline{PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					QueueInfo: nil,
				},
			}},
			want: false,
		},
		{
			p: &Pipeline{PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					QueueInfo: &QueueInfo{
						QueueID:          1,
						EnqueueCondition: apistructs.EnqueueConditionSkipAlreadyRunningLimit,
					},
				},
			}},
			want: true,
		},
		{
			p: &Pipeline{PipelineExtra: PipelineExtra{
				Extra: PipelineExtraInfo{
					QueueInfo: &QueueInfo{
						QueueID:          1,
						EnqueueCondition: "skip",
					},
				},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.p.CanSkipRunningCheck())
	}
}

func Test_canDelete(t *testing.T) {
	type args struct {
		p Pipeline
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
				p: Pipeline{
					PipelineBase: PipelineBase{
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
				p: Pipeline{
					PipelineBase: PipelineBase{
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
				p: Pipeline{
					PipelineBase: PipelineBase{
						Status: apistructs.PipelineStatusFailed,
					},
					PipelineExtra: PipelineExtra{
						Extra: PipelineExtraInfo{
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
				p: Pipeline{
					PipelineBase: PipelineBase{
						Status: apistructs.PipelineStatusFailed,
					},
					PipelineExtra: PipelineExtra{
						Extra: PipelineExtraInfo{
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
				p: Pipeline{
					PipelineBase: PipelineBase{
						Status:  apistructs.PipelineStatusFailed,
						TimeEnd: &[]time.Time{time.Now().Add(-5184555 * time.Second)}[0],
					},
					PipelineExtra: PipelineExtra{
						Extra: PipelineExtraInfo{
							GC: basepb.PipelineGC{
								DatabaseGC: &basepb.PipelineDatabaseGC{
									Finished: &basepb.PipelineDBGCItem{
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
				p: Pipeline{
					PipelineBase: PipelineBase{
						Status:  apistructs.PipelineStatusFailed,
						TimeEnd: &[]time.Time{time.Now().Add(-5182555 * time.Second)}[0],
					},
					PipelineExtra: PipelineExtra{
						Extra: PipelineExtraInfo{
							GC: basepb.PipelineGC{
								DatabaseGC: &basepb.PipelineDatabaseGC{
									Finished: &basepb.PipelineDBGCItem{
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
			got, got1 := tt.args.p.CanDelete()
			if got != tt.want {
				t.Errorf("canDelete() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("canDelete() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetOwnerUserID(t *testing.T) {
	type args struct {
		extra PipelineExtra
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty owner user",
			args: args{
				extra: PipelineExtra{},
			},
			want: "",
		},
		{
			name: "with owner user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						OwnerUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
					},
				},
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				PipelineExtra: tt.args.extra,
			}
			if got := p.GetOwnerUserID(); got != tt.want {
				t.Errorf("GetOwnerUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	type args struct {
		extra PipelineExtra
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty owner user",
			args: args{
				extra: PipelineExtra{},
			},
			want: "",
		},
		{
			name: "with submit user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						SubmitUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
					},
				},
			},
			want: "1",
		},
		{
			name: "with run user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						SubmitUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
						RunUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("2"),
						},
					},
				},
			},
			want: "2",
		},
		{
			name: "with owner user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						SubmitUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
						RunUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("2"),
						},
						OwnerUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("3"),
						},
					},
				},
			},
			want: "3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				PipelineExtra: tt.args.extra,
			}
			if got := p.GetUserID(); got != tt.want {
				t.Errorf("GetOwnerUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOwnerOrRunUserID(t *testing.T) {
	type args struct {
		extra       PipelineExtra
		triggerMode apistructs.PipelineTriggerMode
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with empty owner user",
			args: args{
				extra: PipelineExtra{},
			},
			want: "",
		},
		{
			name: "with owner user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						OwnerUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
					},
				},
			},
			want: "",
		},
		{
			name: "with run user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						RunUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("2"),
						},
					},
				},
			},
			want: "2",
		},
		{
			name: "cron with owner user",
			args: args{
				extra: PipelineExtra{
					Extra: PipelineExtraInfo{
						RunUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("2"),
						},
						OwnerUser: &basepb.PipelineUser{
							ID: structpb.NewStringValue("1"),
						},
					},
				},
				triggerMode: apistructs.PipelineTriggerModeCron,
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				PipelineBase: PipelineBase{
					TriggerMode: tt.args.triggerMode,
				},
				PipelineExtra: tt.args.extra,
			}
			if got := p.GetOwnerOrRunUserID(); got != tt.want {
				t.Errorf("GetOwnerOrRunUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}
