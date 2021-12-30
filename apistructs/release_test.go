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

package apistructs

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func Test_ConvertToQueryParams(t *testing.T) {
	req := ReleaseListRequest{
		ReleaseName: "fake-release",
		Cluster:     "fake-cluster",
	}

	m := req.ConvertToQueryParams()
	assert.Equal(t, m["cluster"], []string{"fake-cluster"})
	assert.Equal(t, m["releaseName"], []string{"fake-release"})
}

const testDiceyml = `
addons: {}
envs: {}
jobs: {}
services:
  doc:
    deployments:
      replicas: 1
    expose:
    - 80
    health_check:
      http:
        duration: 300
        path: /
        port: 80
    image: registry.cn-hangzhou.aliyuncs.com/dspo/docs:i20211229-0001
    ports:
    - 80
    resources:
      cpu: 0.1
      mem: 128
  web:
    deployments:
      replicas: 1
    image: registry.cn-hangzhou.aliyuncs.com/dspo/web:latest
    ports:
    - 80
version: "2.0"
`

func TestReleaseGetResponseData_ReLoadImages(t *testing.T) {
	var data ReleaseGetResponseData
	data.ApplicationReleaseList = []*ApplicationReleaseSummary{{DiceYml: testDiceyml}}
	if err := data.ReLoadImages(); err != nil {
		t.Fatal(err)
	}
	assertServices(t, data.ApplicationReleaseList[0].Services)

	data.Diceyml = testDiceyml
	if err := data.ReLoadImages(); err != nil {
		t.Fatal(err)
	}
	assertServices(t, data.Services)
}

func assertServices(t *testing.T, services []*ServiceImagePair) {
	if len(services) != 2 {
		t.Fatal("services count error")
	}
	for _, service := range services {
		switch {
		case service.ServiceName == "doc" && service.Image == "registry.cn-hangzhou.aliyuncs.com/dspo/docs:i20211229-0001":
		case service.ServiceName == "web" && services[1].Image == "registry.cn-hangzhou.aliyuncs.com/dspo/web:latest":
		default:
			t.Fatalf("service name or image parse error, serviceName: %s, image: %s",
				service.ServiceName, service.Image)
		}
	}
}
