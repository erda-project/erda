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

package endpoints

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
)

func TestUnmarshalApplicationReleaseList(t *testing.T) {
	list := [][]string{{"1"}, {"2"}, {"3"}}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	res, err := unmarshalApplicationReleaseList(string(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != len(res) {
		t.Errorf("test failed, length of res is not expected")
	}
	for i := range list {
		for j := range list[i] {
			if list[i][j] != res[i][j] {
				t.Errorf("test failed, res is not expected")
			}
		}
	}
}

func TestMakeMetadata(t *testing.T) {
	projectRelease := &dbclient.Release{
		Desc:      "testDesc",
		Changelog: "testMarkdown",
		Version:   "testVersion",
	}
	labels := map[string]string{
		"gitBranch":        "testBranch",
		"gitCommitId":      "testCommitId",
		"gitCommitMessage": "testMsg",
		"gitRepo":          "testRepo",
	}
	data, err := json.Marshal(labels)
	if err != nil {
		t.Fatal(err)
	}
	appReleases := [][]dbclient.Release{
		{
			{
				Changelog:       "testMarkdown",
				Labels:          string(data),
				ApplicationName: "testApp",
			},
		},
	}
	releaseMeta := apistructs.ReleaseMetadata{
		Org:       "testOrg",
		Source:    "erda",
		Author:    "testUser",
		Version:   projectRelease.Version,
		Desc:      projectRelease.Desc,
		ChangeLog: projectRelease.Changelog,
		AppList: [][]apistructs.AppMetadata{
			{
				{
					AppName:          "testApp",
					GitBranch:        labels["gitBranch"],
					GitCommitID:      labels["gitCommitId"],
					GitCommitMessage: labels["gitCommitMessage"],
					GitRepo:          labels["gitRepo"],
					ChangeLog:        appReleases[0][0].Changelog,
				},
			},
		},
	}

	target, err := yaml.Marshal(releaseMeta)
	if err != nil {
		t.Fatal(err)
	}
	res, err := makeMetadata("testOrg", "testUser", projectRelease, appReleases)
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != string(target) {
		t.Errorf("test failed, res is not expected")
	}
}
