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

package pipelinesvc

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestPipelineSvc_BatchCreateLabels(t *testing.T) {
	var dbClient *dbclient.Client

	m := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "BatchInsertLabels",
		func(d *dbclient.Client, labels []spec.PipelineLabel, ops ...dbclient.SessionOption) error {
			return nil
		},
	)
	defer m.Unpatch()

	type fields struct {
		dbClient *dbclient.Client
	}
	type args struct {
		createReq *apistructs.PipelineLabelBatchInsertRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test batch create labels",
			fields: fields{dbClient: dbClient},
			args: args{createReq: &apistructs.PipelineLabelBatchInsertRequest{
				Labels: []apistructs.PipelineLabel{
					{
						Type:            "p_i",
						TargetID:        11807139039661,
						PipelineSource:  "erda",
						PipelineYmlName: "pipeline.yml",
						Key:             "foo",
						Value:           "foo",
					},
				},
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PipelineSvc{
				dbClient: tt.fields.dbClient,
			}
			if err := s.BatchCreateLabels(tt.args.createReq); (err != nil) != tt.wantErr {
				t.Errorf("BatchCreateLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
