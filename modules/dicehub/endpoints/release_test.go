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
	"net/http"
	"net/url"
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
)

func TestEndpoints_getListParams(t *testing.T) {
	query := map[string]string{
		"pageSize":                     "20",
		"pageNo":                       "1",
		"cluster":                      "testCluster",
		"branchName":                   "testBranch",
		"isVersion":                    "true",
		"applicationId":                "1",
		"projectId":                    "1",
		"q":                            "test",
		"startTime":                    "1640656800000",
		"endTime":                      "1640744622000",
		"releaseName":                  "testName",
		"crossCluster":                 "true",
		"crossClusterOrSpecifyCluster": "cross",
		"isStable":                     "true",
		"isFormal":                     "true",
		"isProjectRelease":             "true",
		"userId":                       "1",
		"version":                      "1.0",
		"commitId":                     "123456",
		"tags":                         "test",
		"orderBy":                      "testField",
		"order":                        "DESC",
	}

	path := "/api/erda/releases?"
	for k, v := range query {
		path += k + "=" + v + "&"
	}
	path = path[:len(path)-1]

	u, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	r := http.Request{URL: u}
	e := &Endpoints{}
	req, err := e.getListParams(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.PageSize != 20 || req.PageNum != 1 || req.Cluster != query["cluster"] || req.Branch != query["branchName"] ||
		req.IsVersion != true || len(req.ApplicationID) != 1 || req.ApplicationID[0] != query["applicationId"] ||
		req.ProjectID != 1 || req.Query != query["q"] || req.StartTime != 1640656800000 || req.EndTime != 1640744622000 ||
		req.ReleaseName != query["releaseName"] || req.CrossCluster == nil || *req.CrossCluster != true ||
		req.CrossClusterOrSpecifyCluster == nil || *req.CrossClusterOrSpecifyCluster != query["crossClusterOrSpecifyCluster"] ||
		req.IsStable == nil || *req.IsStable != true || req.IsFormal == nil || *req.IsFormal != true ||
		req.IsProjectRelease == nil || *req.IsProjectRelease != true || len(req.UserID) != 1 || req.UserID[0] != query["userId"] ||
		req.Version != query["version"] || req.CommitID != query["commitId"] || req.Tags != query["tags"] ||
		req.OrderBy != query["orderBy"] || req.Order != query["order"] {
		t.Errorf("test failed, req is not expected")
	}
}

func TestUnmarshalApplicationReleaseList(t *testing.T) {
	list := []string{"1", "2", "3"}
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
		if list[i] != res[i] {
			t.Errorf("test failed, res is not expected")
		}
	}
}

func TestMakeMetadata(t *testing.T) {
	projectRelease := &dbclient.Release{
		Desc:     "testDesc",
		Markdown: "testMarkdown",
		Version:  "testVersion",
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
	appReleases := []dbclient.Release{
		{
			Markdown:        "testMarkdown",
			Labels:          string(data),
			ApplicationName: "testApp",
		},
	}
	releaseMeta := apistructs.ReleaseMetadata{
		Version:   projectRelease.Version,
		Desc:      projectRelease.Desc,
		ChangeLog: projectRelease.Markdown,
		AppList: map[string]apistructs.AppMetadata{
			appReleases[0].ApplicationName: {
				GitBranch:        labels["gitBranch"],
				GitCommitID:      labels["gitCommitId"],
				GitCommitMessage: labels["gitCommitMessage"],
				GitRepo:          labels["gitRepo"],
				ChangeLog:        appReleases[0].Markdown,
			},
		},
	}

	target, err := yaml.Marshal(releaseMeta)
	if err != nil {
		t.Fatal(err)
	}
	res, err := makeMetadata(projectRelease, appReleases)
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != string(target) {
		t.Errorf("test failed, res is not expected")
	}
}

func TestParseMetadata(t *testing.T) {
	target := apistructs.ReleaseMetadata{
		Version:   "1.0",
		Desc:      "testDesc",
		ChangeLog: "testChangelog",
		AppList: map[string]apistructs.AppMetadata{
			"test-app": {
				GitBranch:        "feature/1.0",
				GitCommitID:      "12345678",
				GitCommitMessage: "testMsg",
				GitRepo:          "http://test.com/testApp",
				ChangeLog:        "testChangeLog",
			},
		},
	}

	file, err := os.Open("./release_test_data.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	metadata, err := parseMetadata(file)
	if err != nil {
		t.Fatal(err)
	}

	if metadata.Version != target.Version || metadata.Desc != target.Desc || metadata.ChangeLog != target.ChangeLog ||
		metadata.AppList["test-app"] != target.AppList["test-app"] {
		t.Errorf("test failed, result metadata is not expected")
	}
}
