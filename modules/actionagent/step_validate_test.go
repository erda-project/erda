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
