package k8s

import (
	"fmt"

	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/pkg/strutil"
)

func (k *Kubernetes) createDaemonSet(service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	daemonset, err := k.newDaemonSet(service, sg)
	if err != nil {
		return errors.Errorf("failed to generate daemonset struct, name: %s, (%v)", service.Name, err)
	}

	return k.ds.Create(daemonset)
}

func (k *Kubernetes) getDaemonSetStatus(service *apistructs.Service) (apistructs.StatusDesc, error) {
	var statusDesc apistructs.StatusDesc
	dsName := getDeployName(service)
	daemonSet, err := k.getDaemonSet(service.Namespace, dsName)
	if err != nil {
		return statusDesc, err
	}
	status := daemonSet.Status

	statusDesc.Status = apistructs.StatusUnknown

	if status.NumberAvailable == status.DesiredNumberScheduled {
		statusDesc.Status = apistructs.StatusReady
	} else {
		statusDesc.Status = apistructs.StatusUnHealthy
	}

	return statusDesc, nil
}

func (k *Kubernetes) deleteDaemonSet(namespace, name string) error {
	return k.ds.Delete(namespace, name)
}

func (k *Kubernetes) updateDaemonSet(ds *appsv1.DaemonSet) error {
	return k.ds.Update(ds)
}

func (k *Kubernetes) getDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	return k.ds.Get(namespace, name)
}

func (k *Kubernetes) newDaemonSet(service *apistructs.Service, sg *apistructs.ServiceGroup) (*appsv1.DaemonSet, error) {
	deployName := getDeployName(service)
	daemonset := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: service.Namespace,
			Labels:    make(map[string]string),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deployName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deployName,
					Labels: make(map[string]string),
				},
				Spec: corev1.PodSpec{
					EnableServiceLinks:    func(enable bool) *bool { return &enable }(false),
					ShareProcessNamespace: func(b bool) *bool { return &b }(false),
					Tolerations:           toleration.GenTolerations(),
				},
			},
			UpdateStrategy:       appsv1.DaemonSetUpdateStrategy{},
			MinReadySeconds:      0,
			RevisionHistoryLimit: func(i int32) *int32 { return &i }(int32(3)),
		},
		Status: appsv1.DaemonSetStatus{},
	}

	if v := k.options["FORCE_BLUE_GREEN_DEPLOY"]; v != "true" &&
		(strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.DevWorkspace.String() ||
			strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.TestWorkspace.String()) {
		daemonset.Spec.UpdateStrategy = appsv1.DaemonSetUpdateStrategy{Type: appsv1.RollingUpdateDaemonSetStrategyType}
	}

	affinity := constraintbuilders.K8S(&sg.ScheduleInfo2, service, []constraints.PodLabelsForAffinity{
		{PodLabels: map[string]string{"app": deployName}}}, k).Affinity
	daemonset.Spec.Template.Spec.Affinity = &affinity

	// 1核等于1000m
	cpu := fmt.Sprintf("%.fm", service.Resources.Cpu*1000)
	// 1Mi=1024K=1024x1024字节
	memory := fmt.Sprintf("%.fMi", service.Resources.Mem)

	container := corev1.Container{
		// TODO, container name e.g. redis-1528180634
		Name:  deployName,
		Image: service.Image,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(memory),
			},
		},
	}

	//根据环境设置超分比
	cpuSubscribeRatio := k.cpuSubscribeRatio
	memSubscribeRatio := k.memSubscribeRatio
	switch strutil.ToUpper(service.Env["DICE_WORKSPACE"]) {
	case "DEV":
		cpuSubscribeRatio = k.devCpuSubscribeRatio
		memSubscribeRatio = k.devMemSubscribeRatio
	case "TEST":
		cpuSubscribeRatio = k.testCpuSubscribeRatio
		memSubscribeRatio = k.testMemSubscribeRatio
	case "STAGING":
		cpuSubscribeRatio = k.stagingCpuSubscribeRatio
		memSubscribeRatio = k.stagingMemSubscribeRatio
	}

	// 根据超卖比，设置细粒度的CPU
	if err := k.SetFineGrainedCPU(&container, sg.Extra, cpuSubscribeRatio); err != nil {
		return nil, err
	}

	if err := k.SetOverCommitMem(&container, memSubscribeRatio); err != nil {
		return nil, err
	}

	// 生成 sidecars 容器配置
	sidecars := k.generateSidecarContainers(service.SideCars)

	// 生成 initcontainer 配置
	initcontainers := k.generateInitContainer(service.InitContainer)

	containers := []corev1.Container{container}
	containers = append(containers, sidecars...)
	daemonset.Spec.Template.Spec.Containers = containers
	if len(initcontainers) > 0 {
		daemonset.Spec.Template.Spec.InitContainers = initcontainers
	}

	daemonset.Spec.Selector.MatchLabels[LabelRuntimeID] = service.Env[KeyDiceRuntimeID]
	daemonset.Spec.Template.Labels[LabelRuntimeID] = service.Env[KeyDiceRuntimeID]
	daemonset.Labels[LabelRuntimeID] = service.Env[KeyDiceRuntimeID]
	daemonset.Labels["app"] = deployName
	daemonset.Spec.Template.Labels["app"] = deployName

	if daemonset.Spec.Template.Annotations == nil {
		daemonset.Spec.Template.Annotations = make(map[string]string)
	}
	podAnnotations(service, daemonset.Spec.Template.Annotations)

	// 按当前的设定，一个pod中只有一个用户的容器
	if service.Cmd != "" {
		for i := range containers {
			cmds := []string{"sh", "-c", service.Cmd}
			containers[i].Command = cmds
		}
	}
	SetHealthCheck(&daemonset.Spec.Template.Spec.Containers[0], service)

	if err := k.AddContainersEnv(containers, service, sg); err != nil {
		return nil, err
	}

	k.AddSpotEmptyDir(&daemonset.Spec.Template.Spec)

	return daemonset, nil
}
