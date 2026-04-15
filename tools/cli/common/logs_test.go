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

package common

import (
	"testing"
	"time"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFlattenPipelineTasks(t *testing.T) {
	pipeline := testPipelineDetail()

	got := FlattenPipelineTasks(pipeline)
	if len(got) != 3 {
		t.Fatalf("FlattenPipelineTasks() len = %d, want 3", len(got))
	}
	if got[0].StageIndex != 0 || got[0].TaskIndex != 0 || got[0].Name != "build" {
		t.Fatalf("unexpected first task: %#v", got[0])
	}
	if got[2].StageIndex != 1 || got[2].TaskIndex != 0 || got[2].Name != "release" {
		t.Fatalf("unexpected last task: %#v", got[2])
	}
}

func TestSelectTasks(t *testing.T) {
	pipeline := testPipelineDetail()

	tests := []struct {
		name    string
		opts    SelectTasksOptions
		wantIDs []uint64
		wantErr bool
	}{
		{
			name:    "select by task id",
			opts:    SelectTasksOptions{TaskID: 102},
			wantIDs: []uint64{102},
		},
		{
			name:    "select by task name",
			opts:    SelectTasksOptions{TaskName: "build"},
			wantIDs: []uint64{101},
		},
		{
			name:    "select failed tasks",
			opts:    SelectTasksOptions{FailedOnly: true},
			wantIDs: []uint64{102},
		},
		{
			name:    "select running tasks",
			opts:    SelectTasksOptions{RunningOnly: true},
			wantIDs: []uint64{103},
		},
		{
			name:    "select all tasks",
			opts:    SelectTasksOptions{All: true},
			wantIDs: []uint64{101, 102, 103},
		},
		{
			name:    "missing task id fails",
			opts:    SelectTasksOptions{TaskID: 999},
			wantErr: true,
		},
		{
			name:    "missing task name fails",
			opts:    SelectTasksOptions{TaskName: "missing"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectTasks(pipeline, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SelectTasks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("SelectTasks() len = %d, want %d", len(got), len(tt.wantIDs))
			}
			for i := range got {
				if got[i].ID != tt.wantIDs[i] {
					t.Fatalf("SelectTasks()[%d].ID = %d, want %d", i, got[i].ID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestSelectTasksSkipsPlaceholderTasks(t *testing.T) {
	pipeline := testPipelineDetail()
	pipeline.PipelineStages[1].PipelineTasks = append(pipeline.PipelineStages[1].PipelineTasks,
		&basepb.PipelineTaskDTO{
			ID:         0,
			PipelineID: 1000,
			StageID:    12,
			Name:       "dice",
			Status:     apistructs.PipelineStatusAnalyzed.String(),
		},
	)

	got, err := SelectTasks(pipeline, SelectTasksOptions{All: true})
	if err != nil {
		t.Fatalf("SelectTasks() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("SelectTasks() len = %d, want 3", len(got))
	}
	for _, task := range got {
		if task.ID == 0 {
			t.Fatalf("SelectTasks() returned placeholder task: %#v", task)
		}
	}
}

func TestBuildTaskLogQueryParams(t *testing.T) {
	end := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	params := buildTaskLogQueryParams(TaskLogRequestOptions{
		PipelineID:  1000,
		TaskID:      101,
		Tail:        200,
		Stream:      "stdout",
		ClusterName: "erda-cloud",
		LogID:       "pipeline-task-101",
	}, end)

	if params["count"] != "-200" {
		t.Fatalf("count = %q, want -200", params["count"])
	}
	if params["start"] != "0" {
		t.Fatalf("start = %q, want 0", params["start"])
	}
	if params["stream"] != "stdout" {
		t.Fatalf("stream = %q, want stdout", params["stream"])
	}
	if params["clusterName"] != "erda-cloud" {
		t.Fatalf("clusterName = %q, want erda-cloud", params["clusterName"])
	}
	if params["taskID"] != "101" {
		t.Fatalf("taskID = %q, want 101", params["taskID"])
	}
	if params["id"] != "pipeline-task-101" {
		t.Fatalf("id = %q, want pipeline-task-101", params["id"])
	}
}

func TestBuildTaskLogQueryParamsDefaultsLogID(t *testing.T) {
	end := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	params := buildTaskLogQueryParams(TaskLogRequestOptions{
		PipelineID: 1000,
		TaskID:     101,
		Tail:       200,
		Stream:     "stdout",
	}, end)

	if params["id"] != "pipeline-task-101" {
		t.Fatalf("id = %q, want pipeline-task-101", params["id"])
	}
}

func TestBuildTaskLogQueryParamsSupportsExplicitWindow(t *testing.T) {
	end := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	params := buildTaskLogQueryParams(TaskLogRequestOptions{
		PipelineID: 1000,
		TaskID:     101,
		Tail:       200,
		Stream:     "stdout",
		Start:      1776157455467386849,
		End:        1776159394201000000,
		Count:      200,
	}, end)

	if params["count"] != "200" {
		t.Fatalf("count = %q, want 200", params["count"])
	}
	if params["start"] != "1776157455467386849" {
		t.Fatalf("start = %q, want explicit start", params["start"])
	}
	if params["end"] != "1776159394201000000" {
		t.Fatalf("end = %q, want explicit end", params["end"])
	}
}

func testPipelineDetail() pipelinepb.PipelineDetailDTO {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	return pipelinepb.PipelineDetailDTO{
		ID: 1000,
		PipelineStages: []*basepb.PipelineStageDetailDTO{
			{
				ID: 11,
				PipelineTasks: []*basepb.PipelineTaskDTO{
					{
						ID:         101,
						PipelineID: 1000,
						StageID:    11,
						Name:       "build",
						Status:     apistructs.PipelineStatusSuccess.String(),
						TimeBegin:  timestamppb.New(now),
						Extra:      &basepb.PipelineTaskExtra{UUID: "uuid-build"},
					},
					{
						ID:         102,
						PipelineID: 1000,
						StageID:    11,
						Name:       "test",
						Status:     apistructs.PipelineStatusFailed.String(),
						TimeBegin:  timestamppb.New(now.Add(time.Minute)),
						Extra:      &basepb.PipelineTaskExtra{UUID: "uuid-test"},
					},
				},
			},
			{
				ID: 12,
				PipelineTasks: []*basepb.PipelineTaskDTO{
					{
						ID:         103,
						PipelineID: 1000,
						StageID:    12,
						Name:       "release",
						Status:     apistructs.PipelineStatusRunning.String(),
						TimeBegin:  timestamppb.New(now.Add(2 * time.Minute)),
						Extra:      &basepb.PipelineTaskExtra{UUID: "uuid-release"},
					},
				},
			},
		},
	}
}
