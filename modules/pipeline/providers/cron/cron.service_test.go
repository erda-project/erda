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

package cron

import (
	"context"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/xormplus/xorm"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	. "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
)

func Test_provider_CronCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CronCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.CronCreateResponse
		wantErr bool
	}{
		{
			name: "test nil req",
			args: args{
				ctx: context.Background(),
				req: &pb.CronCreateRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test normal req",
			args: args{
				ctx: context.Background(),
				req: &pb.CronCreateRequest{
					PipelineSource:  "test",
					PipelineYmlName: "test",
					PipelineYml: `version: 1.1
cron_compensator:
  enable: true
  latest_first: true
  stop_if_latter_executed: true
stages: []
`,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &provider{}

			patch := monkey.Patch(Transaction, func(dbClient *db.Client, do func(option mysqlxorm.SessionOption) error) error {
				return nil
			})
			defer patch.Unpatch()

			got, err := s.CronCreate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CronCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CronCreate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pbCronToDBCron(t *testing.T) {
	stringTime := "2017-08-30 16:40:41"
	loc, _ := time.LoadLocation("UTC")
	parseTime, _ := time.ParseInLocation("2006-01-02 15:04:05", stringTime, loc)

	type args struct {
		pbCron *Cron
	}
	tests := []struct {
		name       string
		args       args
		wantResult *db.PipelineCron
		wantErr    bool
	}{
		{
			name: "test match",
			args: args{
				pbCron: &Cron{
					ID:                     1,
					TimeCreated:            timestamppb.New(parseTime),
					TimeUpdated:            timestamppb.New(parseTime),
					ApplicationID:          1,
					Branch:                 "test",
					CronExpr:               "test",
					CronStartTime:          timestamppb.New(parseTime),
					PipelineYmlName:        "test",
					BasePipelineID:         1,
					Enable:                 wrapperspb.Bool(true),
					PipelineYml:            "test",
					ConfigManageNamespaces: []string{"1"},
					UserID:                 "test",
					OrgID:                  1,
					PipelineDefinitionID:   "test",
					PipelineSource:         "test",
					Secrets: map[string]string{
						"test": "test",
					},
					Extra: &CronExtra{
						PipelineYml: "test",
						ClusterName: "test",
						Labels: map[string]string{
							"test": "test",
						},
						NormalLabels: map[string]string{
							"test": "test",
						},
						Envs: map[string]string{
							"test": "test",
						},
						ConfigManageNamespaces: []string{"test"},
						IncomingSecrets:        map[string]string{"test": "test"},
						CronStartFrom:          timestamppb.New(parseTime),
						Version:                "v2",
						Compensator: &CronCompensator{
							Enable:               wrapperspb.Bool(true),
							LatestFirst:          wrapperspb.Bool(true),
							StopIfLatterExecuted: wrapperspb.Bool(true),
						},
						LastCompensateAt: timestamppb.New(parseTime),
					},
					IsEdge: wrapperspb.Bool(false),
				},
			},
			wantErr: false,
			wantResult: &db.PipelineCron{
				ID:              1,
				TimeCreated:     parseTime,
				TimeUpdated:     parseTime,
				PipelineSource:  apistructs.PipelineSource("test"),
				PipelineYmlName: "test",
				CronExpr:        "test",
				Enable:          &[]bool{true}[0],
				Extra: db.PipelineCronExtra{
					PipelineYml: "test",
					ClusterName: "test",
					FilterLabels: map[string]string{
						"test": "test",
					},
					NormalLabels: map[string]string{
						"test": "test",
					},
					Envs: map[string]string{
						"test": "test",
					},
					ConfigManageNamespaces: []string{"test"},
					IncomingSecrets: map[string]string{
						"test": "test",
					},
					CronStartFrom: &parseTime,
					Version:       "v2",
					Compensator: &apistructs.CronCompensator{
						Enable:               true,
						LatestFirst:          true,
						StopIfLatterExecuted: true,
					},
					LastCompensateAt: &parseTime,
				},
				ApplicationID:        1,
				Branch:               "test",
				BasePipelineID:       1,
				PipelineDefinitionID: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := pbCronToDBCron(tt.args.pbCron)
			assert.NoError(t, err)
			assert.EqualValues(t, *gotResult, *tt.wantResult)
		})
	}
}

func Test_provider_InsertOrUpdatePipelineCron(t *testing.T) {
	type args struct {
		new_ *db.PipelineCron
		ops  []mysqlxorm.SessionOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test update v1 cron",
			args: args{
				new_: &db.PipelineCron{
					ApplicationID:   1,
					Branch:          "test",
					PipelineYmlName: "test",
				},
			},
		},
		{
			name: "test source and ymlName",
			args: args{
				new_: &db.PipelineCron{
					PipelineSource:  "test",
					PipelineYmlName: "test",
				},
			},
		},
		{
			name: "test create",
			args: args{
				new_: &db.PipelineCron{
					CronExpr: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &provider{}

			var engine xorm.Engine
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(&engine), "Get", func(engine *xorm.Engine, bean interface{}) (bool, error) {
				cron := bean.(*db.PipelineCron)
				assert.Equal(t, cron.ApplicationID, tt.args.new_.ApplicationID)
				assert.Equal(t, cron.Branch, tt.args.new_.Branch)
				assert.Equal(t, cron.PipelineYmlName, tt.args.new_.PipelineYmlName)
				cron.ID = 1

				if cron.ApplicationID > 0 {
					return true, nil
				}
				if cron.PipelineSource != "" {
					return true, nil
				}
				return false, nil
			})
			defer patch.Unpatch()

			var dbClient db.Client
			patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "GetDBClient", func(dbClient *db.Client) *xorm.Engine {
				return &engine
			})
			defer patch2.Unpatch()

			patch3 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "UpdatePipelineCron", func(dbClient *db.Client, id interface{}, cron *db.PipelineCron, ops ...mysqlxorm.SessionOption) error {
				assert.Equal(t, id, uint64(1))
				return nil
			})
			defer patch3.Unpatch()

			patch4 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "CreatePipelineCron", func(bClient *db.Client, cron *db.PipelineCron, ops ...mysqlxorm.SessionOption) error {
				return nil
			})
			defer patch4.Unpatch()

			s.dbClient = &dbClient
			if err := s.InsertOrUpdatePipelineCron(tt.args.new_, tt.args.ops...); (err != nil) != tt.wantErr {
				t.Errorf("InsertOrUpdatePipelineCron() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_provider_disable(t *testing.T) {
	type fields struct{}
	type args struct {
		cron   *db.PipelineCron
		option mysqlxorm.SessionOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test update v1 cron",
			args: args{
				cron: &db.PipelineCron{
					ApplicationID:   1,
					Branch:          "test",
					PipelineYmlName: "test",
				},
			},
		},
		{
			name: "test source and ymlName",
			args: args{
				cron: &db.PipelineCron{
					PipelineSource:  "test",
					PipelineYmlName: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &provider{}

			var engine xorm.Engine
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(&engine), "Get", func(engine *xorm.Engine, bean interface{}) (bool, error) {
				cron := bean.(*db.PipelineCron)
				cron.ID = 1

				if cron.ApplicationID > 0 {
					return true, nil
				}
				if cron.PipelineSource != "" {
					return true, nil
				}
				return false, nil
			})
			defer patch.Unpatch()

			var dbClient db.Client
			patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "GetDBClient", func(dbClient *db.Client) *xorm.Engine {
				return &engine
			})
			defer patch2.Unpatch()

			patch3 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "UpdatePipelineCronWillUseDefault", func(dbClient *db.Client, id interface{}, cron *db.PipelineCron, columns []string, ops ...mysqlxorm.SessionOption) error {
				assert.Equal(t, id, uint64(1))
				return nil
			})
			defer patch3.Unpatch()

			s.dbClient = &dbClient
			if err := s.disable(tt.args.cron, tt.args.option); (err != nil) != tt.wantErr {
				t.Errorf("disable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
