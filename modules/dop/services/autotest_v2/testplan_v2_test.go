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

package autotestv2

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_sceneSetToYml(t *testing.T) {
	type args struct {
		resultsScenes []apistructs.AutoTestScene
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				resultsScenes: []apistructs.AutoTestScene{
					{
						AutoTestSceneParams: apistructs.AutoTestSceneParams{
							ID: 1,
						},
						Inputs: []apistructs.AutoTestSceneInput{
							{
								Name:        "test",
								Description: "test",
								Value:       "test",
							},
						},
						Output: []apistructs.AutoTestSceneOutput{
							{
								Name:        "test",
								Description: "test",
								Value:       "test",
							},
						},
						Steps: []apistructs.AutoTestSceneStep{
							{
								Type:    apistructs.StepTypeScene,
								Method:  apistructs.StepAPIMethodGet,
								Name:    "test",
								SpaceID: 1,
								SceneID: 1,
								AutoTestSceneParams: apistructs.AutoTestSceneParams{
									ID: 1,
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want:    "version: \"1.1\"\nstages:\n  - stage:\n      - snippet:\n          alias: \"1\"\n          params:\n            test: test\n          labels:\n            AUTOTESTTYPE: SCENE\n            SCENE: eyJpZCI6MSwibmFtZSI6IiIsImRlc2NyaXB0aW9uIjoiIiwicHJlSUQiOjAsInNldElEIjowLCJjcmVhdGVBdCI6bnVsbCwidXBkYXRlQXQiOm51bGwsInN0YXR1cyI6IiIsInN0ZXBDb3VudCI6MCwiaW5wdXRzIjpbeyJuYW1lIjoidGVzdCIsImRlc2NyaXB0aW9uIjoidGVzdCIsInZhbHVlIjoidGVzdCIsInRlbXAiOiIiLCJzY2VuZUlEIjowfV0sIm91dHB1dCI6W3sibmFtZSI6InRlc3QiLCJkZXNjcmlwdGlvbiI6InRlc3QiLCJ2YWx1ZSI6InRlc3QiLCJzY2VuZUlEIjowfV0sInN0ZXBzIjpbeyJpZCI6MSwidHlwZSI6IlNDRU5FIiwibWV0aG9kIjoiR0VUIiwidmFsdWUiOiIiLCJuYW1lIjoidGVzdCIsInByZUlEIjowLCJwcmVUeXBlIjoiIiwic2NlbmVJRCI6MSwic3BhY2VJRCI6MSwiY3JlYXRvcklEIjoiIiwidXBkYXRlcklEIjoiIiwiQ2hpbGRyZW4iOm51bGwsImFwaVNwZWNJRCI6MH1dLCJyZWZTZXRJRCI6MH0=\n          snippet_config:\n            source: autotest\n            name: \"1\"\n            labels:\n              autotestExecType: scene\n              sceneID: \"1\"\n              spaceID: \"0\"\n          if: ${{ 1 == 1 }}\noutputs:\n  - name: 1_test\n    ref: ${{ outputs.1.test }}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sceneSetToYml(tt.args.resultsScenes)
			if (err != nil) != tt.wantErr {
				t.Errorf("sceneSetToYml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sceneSetToYml() got = %v, want %v", got, tt.want)
			}
		})
	}
}
