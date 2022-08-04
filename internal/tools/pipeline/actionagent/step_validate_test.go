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

package actionagent

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentConvertEnvsByClusterLocation(t *testing.T) {
	os.Clearenv()
	for _, kv := range os.Environ() {
		fmt.Println(kv)
	}
	agent := &Agent{
		Arg: &AgentArg{},
	}

	os.Setenv("OPENAPI_ADDR", "openapi addr")
	os.Setenv("OPENAPI_PUBLIC_URL", "openapi public url")

	os.Setenv("SONAR_ADDR", "sonar addr")
	os.Setenv("SONAR_PUBLIC_URL", "sonar public url")

	// edge cluster
	agent.EasyUse.IsEdgeCluster = true

	agent.convertEnvsByClusterLocation()
	fmt.Println("---> edge cluster operate")
	assert.Equal(t, 0, len(agent.Errs))

	for _, kv := range os.Environ() {
		fmt.Println(kv)
	}
	assert.Equal(t, "openapi public url", os.Getenv("OPENAPI_ADDR"))
	assert.Equal(t, "openapi public url", os.Getenv("OPENAPI_PUBLIC_URL"))

	assert.Equal(t, "sonar public url", os.Getenv("SONAR_ADDR"))
	assert.Equal(t, "sonar public url", os.Getenv("SONAR_PUBLIC_URL"))
}

func TestAgentCentralClusterConvertEnvsByClusterLocation2(t *testing.T) {
	os.Clearenv()
	for _, kv := range os.Environ() {
		fmt.Println(kv)
	}
	agent := &Agent{
		Arg: &AgentArg{},
	}

	os.Setenv("OPENAPI_ADDR", "openapi addr")
	os.Setenv("OPENAPI_PUBLIC_URL", "openapi public url")

	os.Setenv("SONAR_ADDR", "sonar addr")
	os.Setenv("SONAR_PUBLIC_URL", "sonar public url")

	os.Setenv("XXX_PUBLIC_URL", "xxx public url")
	os.Setenv("XXX_ADDR", "")

	os.Setenv("YYY_PUBLIC_URL", "yyy public url")
	// os.Setenv("YYY_ADDR", "")

	// central cluster
	agent.EasyUse.IsEdgeCluster = false

	agent.convertEnvsByClusterLocation()
	fmt.Println("---> central cluster operate")
	assert.Equal(t, 0, len(agent.Errs))

	for _, kv := range os.Environ() {
		fmt.Println(kv)
	}
	assert.Equal(t, "openapi addr", os.Getenv("OPENAPI_ADDR"))
	assert.Equal(t, "openapi public url", os.Getenv("OPENAPI_PUBLIC_URL"))

	assert.Equal(t, "sonar addr", os.Getenv("SONAR_ADDR"))
	assert.Equal(t, "sonar public url", os.Getenv("SONAR_PUBLIC_URL"))

	assert.Equal(t, "", os.Getenv("XXX_ADDR"))
	assert.Equal(t, "xxx public url", os.Getenv("XXX_PUBLIC_URL"))

	assert.Equal(t, "yyy public url", os.Getenv("YYY_ADDR"))
	assert.Equal(t, "yyy public url", os.Getenv("YYY_PUBLIC_URL"))
}

func TestSetOpenApiToken(t *testing.T) {
	reporter := &CenterCallbackReporter{}
	agent := &Agent{
		CallbackReporter: reporter,
	}
	agent.CallbackReporter.SetOpenApiToken("token")
	assert.Equal(t, "token", reporter.OpenAPIToken)
}

func TestSetCollectorAddress(t *testing.T) {
	reporter := &CenterCallbackReporter{}
	agent := &Agent{
		CallbackReporter: reporter,
	}
	agent.CallbackReporter.SetCollectorAddress("addr")
	assert.Equal(t, "addr", reporter.CollectorAddr)
}

func Test_validate(t *testing.T) {
	os.Setenv(EnvContextDir, "contextdir")
	os.Setenv(EnvWorkDir, "workdir")
	os.Setenv(EnvMetaFile, "metafile")
	os.Setenv(EnvUploadDir, "uploaddir")
	agent := &Agent{
		Arg: &AgentArg{},
	}
	agent.validate()
	assert.Equal(t, "contextdir", agent.EasyUse.ContainerContext)
	assert.Equal(t, "workdir", agent.EasyUse.ContainerWd)
	assert.Equal(t, "metafile", agent.EasyUse.ContainerMetaFile)
	assert.Equal(t, "uploaddir", agent.EasyUse.ContainerUploadDir)
}

func Test_setShellAndArgs(t *testing.T) {
	type arg struct {
		shell string
	}
	tests := []struct {
		name      string
		arg       arg
		wantShell string
		wantArgs  []string
	}{
		{
			name: "sh shell",
			arg: arg{
				shell: "/bin/sh",
			},
			wantShell: "/bin/sh",
			wantArgs:  []string{},
		},
		{
			name: "sh shell with args",
			arg: arg{
				shell: "/bin/sh -c",
			},
			wantShell: "/bin/sh",
			wantArgs:  []string{"-c"},
		},
		{
			name: "python shell",
			arg: arg{
				shell: "python",
			},
			wantShell: "python",
			wantArgs:  []string{},
		},
	}
	for _, tt := range tests {
		agent := &Agent{
			Arg: &AgentArg{
				Shell: tt.arg.shell,
			},
		}
		agent.setShellAndArgs()
		assert.Equal(t, tt.wantShell, agent.Arg.Shell)
		assert.Equal(t, tt.wantArgs, agent.Arg.ShellArgs)
	}
}
