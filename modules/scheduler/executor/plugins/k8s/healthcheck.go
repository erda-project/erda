package k8s

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
)

func (k *Kubernetes) NewHealthcheckProbe(service *apistructs.Service) *apiv1.Probe {
	return FillHealthCheckProbe(service)
}

func SetHealthCheck(container *apiv1.Container, service *apistructs.Service) {

	probe := FillHealthCheckProbe(service)

	container.LivenessProbe = probe
	readinessprobe := probe.DeepCopy()
	if readinessprobe != nil {
		readinessprobe.FailureThreshold = 3
		readinessprobe.PeriodSeconds = 10
		readinessprobe.InitialDelaySeconds = 10
	}
	container.ReadinessProbe = readinessprobe

}

// FillHealthCheckProbe 基于 service 填充出 k8s probe
func FillHealthCheckProbe(service *apistructs.Service) *apiv1.Probe {
	var (
		probe *apiv1.Probe
		newHC = service.NewHealthCheck
		oldHC = service.HealthCheck
	)

	if newHC != nil && (newHC.ExecHealthCheck != nil || newHC.HttpHealthCheck != nil) {
		probe = NewHealthCheck(newHC)
	} else if oldHC != nil {
		probe = OldHealthCheck(oldHC)
	} else {
		// 默认健康检测
		probe = DefaultHealthCheck(service)
	}

	return probe
}

// NewCheckProbe 创建 k8s probe 默认对象
func NewCheckProbe() *apiv1.Probe {
	return &apiv1.Probe{
		InitialDelaySeconds: 0,
		// 每次健康检查超时时间
		TimeoutSeconds: 10,
		// 健康检查探测间隔
		PeriodSeconds:    15,
		FailureThreshold: int32(apistructs.HealthCheckDuration) / 15,
	}
}

// DefaultHealthCheck 用户没有配置任何健康检查，默认对第一个端口做4层 tcp 检查
func DefaultHealthCheck(service *apistructs.Service) *apiv1.Probe {
	if len(service.Ports) == 0 {
		return nil
	}

	probe := NewCheckProbe()
	probe.TCPSocket = &apiv1.TCPSocketAction{
		Port: intstr.FromInt(service.Ports[0].Port),
	}
	return probe
}

// NewHealthCheck 配置 Dice 新版健康检测
func NewHealthCheck(hc *apistructs.NewHealthCheck) *apiv1.Probe {
	if hc == nil || (hc.HttpHealthCheck == nil && hc.ExecHealthCheck == nil) {
		return nil
	}

	probe := NewCheckProbe()
	if hc.HttpHealthCheck != nil {
		httpCheck := hc.HttpHealthCheck
		probe.HTTPGet = &apiv1.HTTPGetAction{
			Path:   httpCheck.Path,
			Port:   intstr.FromInt(httpCheck.Port),
			Scheme: apiv1.URIScheme("HTTP"),
		}

		if times := int32(httpCheck.Duration) / 15; times > probe.FailureThreshold {
			probe.FailureThreshold = times
		}
	} else if hc.ExecHealthCheck != nil {
		execCheck := hc.ExecHealthCheck
		probe.Exec = &apiv1.ExecAction{
			Command: []string{"sh", "-c", execCheck.Cmd},
		}
		if times := int32(execCheck.Duration) / 15; times > probe.FailureThreshold {
			probe.FailureThreshold = times
		}
	}
	return probe
}

// OldHealthCheck 兼容 Dice 老版本健康检测
func OldHealthCheck(hc *apistructs.HealthCheck) *apiv1.Probe {
	if hc == nil {
		return nil
	}

	probe := NewCheckProbe()
	switch hc.Kind {
	case "COMMAND":
		probe.Exec = &apiv1.ExecAction{
			Command: []string{"sh", "-c", hc.Command},
		}
	case "TCP":
		probe.TCPSocket = &apiv1.TCPSocketAction{
			Port: intstr.FromInt(hc.Port),
		}
	case "HTTP", "https":
		probe.HTTPGet = &apiv1.HTTPGetAction{
			Path:   hc.Path,
			Port:   intstr.FromInt(hc.Port),
			Scheme: apiv1.URIScheme(hc.Kind),
		}
	}
	// 老的健康检查默认持续时长5分钟（连续5分钟内健康检查全部失败则杀死容器），同 dcos 配置
	probe.FailureThreshold = 300 / 15
	return probe
}
