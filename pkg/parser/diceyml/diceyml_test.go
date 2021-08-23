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

package diceyml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var jobyml = `version: 2.0
jobs:
  job1:
    cmd: ls
    envs:
      env1: v1
  job2:
    cmd: ls -l
    envs:
      env2: v2
`

var yml = `version: 2.0

version: 2
envs:
  TERMINUS_APP_NAME: "TEST-global"
  TEST_PARAM: "param_value"
services:
  web:
    ports:
      - 8080
      - port: 20880
      - port: 1234
        protocol: "UDP"
      - port: 4321
        protocol: "HTTP"
      - port: 53
        protocol: "DNS"
        l4_protocol: "UDP"
        default: true
    k8s_snippet:
      container:
        stdin: true
        workingDir: aaa
        imagePullPolicy: Always
        securityContext:
          privileged: true
    health_check:
      exec:
        cmd: "echo 1"
    deployments:
      replicas: ${replicas}
    resources:
      cpu: ${cpu:0.1}
      mem: 512
      disk: 0
    expose:
      - 20880
    volumes:
      - storage: "nfs"
        path: "/data/file/resource"
addons:
  terminus-elasticsearch:
    plan: "terminus-elasticsearch:professional"
    options:
      version: "6.8.9"
  apigateway:
    plan: "api-gateway:basic"
    options:
      version: "2.0.0"
  xxx:
    plan: ${mysql-plan:"mysql:basic"}
    options:
      version: "5.7.23"
values:
  test:
    replicas: 1
    cpu: 0.5
  production:
    replicas: 2
    cpu: 1
    mysql-plan: "rds:basic"
  
`

var wrongSnippetYml = `version: 2.0
services:
  web:
    ports:
      - 8080
      - port: 20880
      - port: 1234
        protocol: "UDP"
      - port: 4321
        protocol: "HTTP"
      - port: 53
        protocol: "DNS"
        l4_protocol: "UDP"
        default: true
    deployments:
      replicas: 1
    resources:
      cpu: 0.1
      mem: 512
    k8s_snippet:
      container:
        name: abc
        stdin: true
        workingDir: aaa
        imagePullPolicy: Always
        securityContext:
          privileged: true
`

func TestDiceYmlObj(t *testing.T) {
	d, err := New([]byte(yml), true)
	assert.Nil(t, err)
	obj := d.Obj()
	assert.Equal(t, "TCP", string(obj.Services["web"].Ports[0].Protocol))
	assert.Equal(t, "TCP", string(obj.Services["web"].Ports[0].L4Protocol))
	assert.Equal(t, "TCP", string(obj.Services["web"].Ports[1].Protocol))
	assert.Equal(t, "TCP", string(obj.Services["web"].Ports[1].L4Protocol))
	assert.Equal(t, "UDP", string(obj.Services["web"].Ports[2].Protocol))
	assert.Equal(t, "UDP", string(obj.Services["web"].Ports[2].L4Protocol))
	assert.Equal(t, "HTTP", string(obj.Services["web"].Ports[3].Protocol))
	assert.Equal(t, "TCP", string(obj.Services["web"].Ports[3].L4Protocol))
	assert.Equal(t, "DNS", string(obj.Services["web"].Ports[4].Protocol))
	assert.Equal(t, "UDP", string(obj.Services["web"].Ports[4].L4Protocol))
	assert.Equal(t, true, obj.Services["web"].Ports[4].Default)
}

func TestDicalYmlK8SSnippet(t *testing.T) {
	d, err := New([]byte(yml), true)
	assert.Nil(t, err)
	assert.NotNil(t, d.obj.Services["web"].K8SSnippet)
	assert.NotNil(t, d.obj.Services["web"].K8SSnippet.Container.SecurityContext)
	assert.Equal(t, true, *d.obj.Services["web"].K8SSnippet.Container.SecurityContext.Privileged)
	assert.EqualValues(t, "Always", d.obj.Services["web"].K8SSnippet.Container.ImagePullPolicy)
	_, err = New([]byte(wrongSnippetYml), true)
	assert.NotNil(t, err)
}

func TestDiceYmlFieldnameValidate(t *testing.T) {
	_, err := New([]byte(yml), true)
	fmt.Printf("%+v\n", err) // debug print
	assert.Nil(t, err)
}

func TestDiceYmlInsertJobImage(t *testing.T) {
	d, err := New([]byte(jobyml), false)
	err = d.InsertImage(map[string]string{"job1": "image1"}, nil)
	assert.Nil(t, err, "%v", err)
	fmt.Printf("%+v\n", d.Obj().Jobs["job1"]) // debug print

}

//func TestDiceYmlMergeValues(t *testing.T) {
//	d, err := NewDeployable([]byte(yml), "test", true)
//	assert.Nil(t, err)
//	assert.Equal(t, 0.5, d.Obj().Services["web"].Resources.CPU)
//	assert.Equal(t, 1, d.Obj().Services["web"].Deployments.Replicas)
//	assert.Equal(t, "mysql:basic", d.Obj().AddOns["xxx"].Plan)
//	d, err = NewDeployable([]byte(yml), "prod", true)
//	assert.Nil(t, err)
//	assert.Equal(t, float64(1), d.Obj().Services["web"].Resources.CPU)
//	assert.Equal(t, 2, d.Obj().Services["web"].Deployments.Replicas)
//	assert.Equal(t, "rds:basic", d.Obj().AddOns["xxx"].Plan)
//}
