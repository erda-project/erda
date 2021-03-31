package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
)

func TestFillHealthCheckProbe(t *testing.T) {
	var probe *corev1.Probe
	service := &apistructs.Service{}

	// nil
	probe = FillHealthCheckProbe(service)
	assert.Nil(t, probe)

	// old hc tcp
	service.HealthCheck = &apistructs.HealthCheck{
		Kind: "TCP",
		Port: 80,
	}
	probe = FillHealthCheckProbe(service)
	assert.NotNil(t, probe)
	assert.Equal(t, service.HealthCheck.Port, probe.TCPSocket.Port.IntValue())

	// old hc http
	service.HealthCheck = &apistructs.HealthCheck{
		Kind: "HTTP",
		Port: 80,
		Path: "/health",
	}
	probe = FillHealthCheckProbe(service)
	assert.NotNil(t, probe)
	assert.Equal(t, service.HealthCheck.Kind, string(probe.HTTPGet.Scheme))
	assert.Equal(t, service.HealthCheck.Path, probe.HTTPGet.Path)
	assert.Equal(t, service.HealthCheck.Port, probe.HTTPGet.Port.IntValue())

	// old hc command
	service.HealthCheck = &apistructs.HealthCheck{
		Kind:    "COMMAND",
		Command: "echo 1",
	}
	probe = FillHealthCheckProbe(service)
	assert.NotNil(t, probe)
	assert.Equal(t, []string{"sh", "-c", service.HealthCheck.Command}, probe.Exec.Command)

	// new hc http
	service.NewHealthCheck = &apistructs.NewHealthCheck{
		HttpHealthCheck: &apistructs.HttpHealthCheck{
			Port:     80,
			Path:     "/health",
			Duration: 1000,
		},
	}
	probe = FillHealthCheckProbe(service)
	assert.NotNil(t, probe)
	assert.Equal(t, service.NewHealthCheck.HttpHealthCheck.Port, probe.HTTPGet.Port.IntValue())
	assert.Equal(t, service.NewHealthCheck.HttpHealthCheck.Path, probe.HTTPGet.Path)
	assert.Equal(t, int32(service.NewHealthCheck.HttpHealthCheck.Duration/15), probe.FailureThreshold)

	// new hc exec
	service.NewHealthCheck = &apistructs.NewHealthCheck{
		ExecHealthCheck: &apistructs.ExecHealthCheck{
			Cmd:      "echo 1",
			Duration: 1000,
		},
	}
	probe = FillHealthCheckProbe(service)
	assert.NotNil(t, probe)
	assert.Equal(t, []string{"sh", "-c", service.NewHealthCheck.ExecHealthCheck.Cmd}, probe.Exec.Command)
	assert.Equal(t, int32(service.NewHealthCheck.ExecHealthCheck.Duration/15), probe.FailureThreshold)
}
