// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package reconciler

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/bmizerany/assert"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestReconciler_getNeedGCPipeline(t *testing.T) {
	type args struct {
		pipelines []spec.Pipeline
		total     int64
		err       error
	}

	var gcTime = func() uint64 {
		return 1800
	}

	now := time.Now()
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "test_empty",
			args: args{
				pipelines: nil,
				total:     0,
				err:       nil,
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "test_return_error",
			args: args{
				pipelines: nil,
				total:     0,
				err:       fmt.Errorf("have error"),
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "completeReconcilerGC_pipeline",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusSuccess,
							TimeEnd: func() *time.Time {
								now := now.Add(-190*time.Second - bufferTime*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       true,
								CompleteReconcilerTeardown: false,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{200}[0],
										SuccessTTLSecond: &[]uint64{200}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "completeReconcilerTeardown_pipeline",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusSuccess,
							TimeEnd: func() *time.Time {
								now := time.Now().Add(-190*time.Second - bufferTime*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: true,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{200}[0],
										SuccessTTLSecond: &[]uint64{200}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "time_over_gc",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusSuccess,
							TimeEnd: func() *time.Time {
								addTime := now.Add(-300*time.Second - bufferTime*time.Second)
								return &addTime
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: false,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{200}[0],
										SuccessTTLSecond: &[]uint64{200}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "time_not_over_gc",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusFailed,
							TimeEnd: func() *time.Time {
								now := now.Add(-190*time.Second - bufferTime*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: false,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{200}[0],
										SuccessTTLSecond: &[]uint64{200}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "over_default_gc_time",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusStopByUser,
							TimeEnd: func() *time.Time {
								now := now.Add(-time.Duration(gcTime())*time.Second - bufferTime*time.Second - 20*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: false,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{gcTime()}[0],
										SuccessTTLSecond: &[]uint64{gcTime()}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "not_over_default_gc_time",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusStopByUser,
							TimeEnd: func() *time.Time {
								now := now.Add(-time.Duration(gcTime())*time.Second - bufferTime*time.Second + 20*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: false,
								GC: apistructs.PipelineGC{
									ResourceGC: apistructs.PipelineResourceGC{
										FailedTTLSecond:  &[]uint64{gcTime()}[0],
										SuccessTTLSecond: &[]uint64{gcTime()}[0],
									},
								},
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "not_end_status",
			args: args{
				pipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							Status: apistructs.PipelineStatusAnalyzed,
							TimeEnd: func() *time.Time {
								now := now.Add(-time.Duration(gcTime())*time.Second - 20*time.Second - bufferTime*time.Second)
								return &now
							}(),
						},
						PipelineExtra: spec.PipelineExtra{
							Extra: spec.PipelineExtraInfo{
								CompleteReconcilerGC:       false,
								CompleteReconcilerTeardown: false,
							},
						},
					},
				},
				total: 11,
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	logrus.Infof("start test pipeline_gc_compensator_test")
	defer logrus.Infof("end test pipeline_gc_compensator_test")

	for _, tt := range tests {

		var db *dbclient.Client
		monkey.PatchInstanceMethod(reflect.TypeOf(db), "PageListPipelines", func(client *dbclient.Client, req apistructs.PipelinePageListRequest, ops ...dbclient.SessionOption) ([]spec.Pipeline, []uint64, int64, int64, error) {
			return tt.args.pipelines, nil, tt.args.total, 0, tt.args.err
		})

		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				dbClient: db,
			}
			got, got1, err := r.getNeedGCPipelines(0, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNeedGCPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantLen, len(got))

			if got1 != tt.args.total {
				t.Errorf("getNeedGCPipeline() got1 = %v, want %v", got1, tt.args.total)
			}
		})

	}
}
