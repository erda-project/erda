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

package docker

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

func TestDockerJobStatus(t *testing.T) {
	tests := []struct {
		name         string
		dockerStatus string
		exitCode     int
		want         apistructs.PipelineStatus
	}{
		{
			name:         "running-task",
			dockerStatus: "running",
			exitCode:     0,
			want:         apistructs.PipelineStatusRunning,
		},
		{
			name:         "success-task",
			dockerStatus: "exited",
			exitCode:     0,
			want:         apistructs.PipelineStatusSuccess,
		},
		{
			name:         "failure-task",
			dockerStatus: "exited",
			exitCode:     1,
			want:         apistructs.PipelineStatusFailed,
		},
		{
			name:         "error-task",
			dockerStatus: "exited",
			exitCode:     127,
			want:         apistructs.PipelineStatusFailed,
		},
		{
			name:         "dead-task",
			dockerStatus: "dead",
			exitCode:     128,
			want:         apistructs.PipelineStatusFailed,
		},
		{
			name:         "unknown-task",
			dockerStatus: "unknown",
			exitCode:     129,
			want:         apistructs.PipelineStatusUnknown,
		},
	}
	d := DockerJob{client: &client.Client{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(d.client), "ContainerInspect", func(cli *client.Client, ctx context.Context, containerID string) (types.ContainerJSON, error) {
				return types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{State: &types.ContainerState{
					Status:   tt.dockerStatus,
					ExitCode: tt.exitCode,
				}}}, nil
			})
			if got, _ := d.Status(context.Background(), &spec.PipelineTask{Extra: spec.PipelineTaskExtra{Namespace: "pipeline-1", UUID: "pipeline-task-1"}}); got.Status != tt.want {
				t.Errorf("DockerJobStatus() = %v, want %v", got, tt.want)
			}
			patch.Unpatch()
		})
	}
}
