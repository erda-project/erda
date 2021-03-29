package pipelinesvc

import (
	"reflect"
	"testing"
	"time"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_getCompensateFromTime(t *testing.T) {

	now := time.Unix(time.Now().Unix(), 0)
	nowSub24h := now.Add(-time.Hour * 24)

	var (
		lastCompensateAtLessThan24h = time.Now().Add(-time.Hour * 2)
		lastCompensateAtMoreThan24h = time.Now().Add(-time.Hour * 25)

		updatedTimeLessThan24h = lastCompensateAtLessThan24h
		updatedTimeMoreThan24h = lastCompensateAtMoreThan24h
	)

	type args struct {
		pc spec.PipelineCron
	}
	tests := []struct {
		name  string
		args  args
		wantT time.Time
	}{
		{
			name: "has lastCompensateAt <24h, use lastCompensateAt",
			args: args{
				pc: spec.PipelineCron{
					Extra: spec.PipelineCronExtra{
						LastCompensateAt: &lastCompensateAtLessThan24h,
					},
				}},
			wantT: lastCompensateAtLessThan24h,
		},
		{
			name: "has lastCompensateAt >24h, use now-24h",
			args: args{
				pc: spec.PipelineCron{
					Extra: spec.PipelineCronExtra{
						LastCompensateAt: &lastCompensateAtMoreThan24h,
					},
				},
			},
			wantT: nowSub24h,
		},
		{
			name: "no lastCompensateAt, timeUpdated <24h, use timeUpdated",
			args: args{
				pc: spec.PipelineCron{
					TimeUpdated: updatedTimeLessThan24h,
				},
			},
			wantT: updatedTimeLessThan24h,
		},
		{
			name: "no lastCompensateAt, timeUpdated >24h, use now-24h",
			args: args{
				pc: spec.PipelineCron{
					TimeUpdated: updatedTimeMoreThan24h,
				},
			},
			wantT: nowSub24h,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotT := getCompensateFromTime(tt.args.pc); !reflect.DeepEqual(gotT, tt.wantT) {
				t.Errorf("getCompensateFromTime() = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}

func Test_orderByCronTriggerTime(t *testing.T) {

	p14 := spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				CronTriggerTime: &[]time.Time{time.Date(2020, 3, 16, 14, 0, 0, 0, time.UTC)}[0],
			},
		},
	}
	p15 := spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				CronTriggerTime: &[]time.Time{time.Date(2020, 3, 16, 15, 0, 0, 0, time.UTC)}[0],
			},
		},
	}
	p16 := spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				CronTriggerTime: &[]time.Time{time.Date(2020, 3, 16, 16, 0, 0, 0, time.UTC)}[0],
			},
		},
	}
	pNULL := spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				CronTriggerTime: nil,
			},
		},
	}
	allP := []spec.Pipeline{p14, p15, p16, pNULL}

	type args struct {
		inputs []spec.Pipeline
		order  string
	}
	tests := []struct {
		name string
		args args
		want []spec.Pipeline
	}{
		{
			name: "desc",
			args: args{
				inputs: allP,
				order:  "DESC",
			},
			want: []spec.Pipeline{p16, p15, p14},
		},
		{
			name: "asc",
			args: args{
				inputs: allP,
				order:  "ASC",
			},
			want: []spec.Pipeline{p14, p15, p16},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderByCronTriggerTime(tt.args.inputs, tt.args.order); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderByCronTriggerTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
