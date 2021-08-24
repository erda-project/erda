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

package k8s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// DefaultServiceDNSSuffix k8s service dns fixed suffix
	DefaultServiceDNSSuffix = "svc.cluster.local"
	shardDirSuffix          = "-shard-dir"
	sidecarNamePrefix       = "sidecar-"
	EnableServiceLinks      = "ENABLE_SERVICE_LINKS"
)

func (k *Kubernetes) createDeployment(service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	deployment, err := k.newDeployment(service, sg)
	if err != nil {
		return errors.Errorf("failed to generate deployment struct, name: %s, (%v)", service.Name, err)
	}

	err = k.deploy.Create(deployment)
	if err != nil {
		return errors.Errorf("failed to create deployment, name: %s, (%v)", service.Name, err)
	}
	if service.K8SSnippet == nil || service.K8SSnippet.Container == nil {
		return nil
	}
	err = k.deploy.Patch(deployment.Namespace, deployment.Name, service.Name, (apiv1.Container)(*service.K8SSnippet.Container))
	if err != nil {
		return errors.Errorf("failed to patch deployment, name: %s, snippet: %+v, (%v)", service.Name, *service.K8SSnippet.Container, err)
	}
	return nil
}

func (k *Kubernetes) getDeploymentStatusFromMap(service *apistructs.Service, deployments map[string]appsv1.Deployment) (apistructs.StatusDesc, error) {
	var (
		statusDesc apistructs.StatusDesc
	)
	// in version 1.10.3, the following two apis are equal
	// http://localhost:8080/apis/extensions/v1beta1/namespaces/default/deployments/myk8stest6
	// http://localhost:8080/apis/extensions/v1beta1/namespaces/default/deployments/myk8stest6/status
	deploymentName := getDeployName(service)

	if deployment, ok := deployments[deploymentName]; ok {

		status := deployment.Status
		//You may get this status when you just start creating
		if len(status.Conditions) == 0 {
			statusDesc.Status = apistructs.StatusUnknown
			statusDesc.LastMessage = "cannot get statusDesc condition"
			return statusDesc, nil
		}

		for _, c := range status.Conditions {
			if c.Type == k8sapi.DeploymentReplicaFailure && c.Status == "True" {
				statusDesc.Status = apistructs.StatusFailing
				return statusDesc, nil
			}
			if c.Type == k8sapi.DeploymentAvailable && c.Status == "False" {
				statusDesc.Status = apistructs.StatusFailing
				return statusDesc, nil
			}
		}

		statusDesc.Status = apistructs.StatusUnknown

		if status.Replicas == status.ReadyReplicas &&
			status.Replicas == status.AvailableReplicas &&
			status.Replicas == status.UpdatedReplicas {
			if status.Replicas > 0 {
				statusDesc.Status = apistructs.StatusReady
				statusDesc.LastMessage = fmt.Sprintf("deployment(%s) is running", deployment.Name)
			} else if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
				statusDesc.Status = apistructs.StatusReady
			} else {
				statusDesc.LastMessage = fmt.Sprintf("deployment(%s) replica is 0, been deleting", deployment.Name)
			}
		}
	}

	return statusDesc, nil
}

func (k *Kubernetes) getDeployment(namespace, name string) (*appsv1.Deployment, error) {
	return k.deploy.Get(namespace, name)
}

func (k *Kubernetes) putDeployment(deployment *appsv1.Deployment, service *apistructs.Service) error {
	err := k.deploy.Put(deployment)
	if err != nil {
		return errors.Errorf("failed to update deployment, name: %s, (%v)", service.Name, err)
	}
	if service.K8SSnippet == nil || service.K8SSnippet.Container == nil {
		return nil
	}
	err = k.deploy.Patch(deployment.Namespace, deployment.Name, service.Name, (apiv1.Container)(*service.K8SSnippet.Container))
	if err != nil {
		return errors.Errorf("failed to patch deployment, name: %s, snippet: %+v, (%v)", service.Name, *service.K8SSnippet.Container, err)
	}
	return nil
}

func (k *Kubernetes) deleteDeployment(namespace, name string) error {
	logrus.Debugf("delete deployment %s on namespace %s", name, namespace)
	return k.deploy.Delete(namespace, name)
}

// AddContainersEnv Add container environment variables
func (k *Kubernetes) AddContainersEnv(containers []apiv1.Container, service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	var envs []apiv1.EnvVar
	ns := MakeNamespace(sg)
	serviceName := service.Name
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		serviceName = service.ProjectServiceName
	}

	// User-injected environment variables
	if len(service.Env) > 0 {
		for k, v := range service.Env {
			envs = append(envs, apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	addEnv := func(svc *apistructs.Service, envs *[]apiv1.EnvVar, useClusterIP bool) error {
		var err error
		// use SHORT dns, service's name is equal to SHORT dns
		svcName := svc.Name
		if svc.ProjectServiceName != "" {
			svcName = svc.ProjectServiceName
		}
		host := strings.Join([]string{svcName, svc.Namespace, DefaultServiceDNSSuffix}, ".")
		if useClusterIP {
			host, err = k.getClusterIP(svc.Namespace, svc.Name)
			if err != nil {
				return err
			}
		}
		// add {serviceName}_HOST
		*envs = append(*envs, apiv1.EnvVar{
			Name:  makeEnvVariableName(svc.Name) + "_HOST",
			Value: host,
		})

		// {serviceName}_PORT Refers to the first port
		if len(svc.Ports) > 0 {
			*envs = append(*envs, apiv1.EnvVar{
				Name:  makeEnvVariableName(svc.Name) + "_PORT",
				Value: strconv.Itoa(svc.Ports[0].Port),
			})
		}

		//If there are multiple ports, use them in sequence: {serviceName}_PORT0,{serviceName}_PORT1,...
		for i, port := range svc.Ports {
			*envs = append(*envs, apiv1.EnvVar{
				Name:  makeEnvVariableName(svc.Name) + "_PORT" + strconv.Itoa(i),
				Value: strconv.Itoa(port.Port),
			})
		}
		return nil
	}

	//All containers have environment variables for all services
	if sg.ServiceDiscoveryMode == "GLOBAL" {
		// Since some services may not be created yet, so inject dns
		// TODO: You can also create the k8s service of all services in the runtime first
		for _, svc := range sg.Services {
			// Services without port exposure have no environment variables such as {serviceName}_HOST, {serviceName}_PORT, etc.
			if len(svc.Ports) == 0 {
				continue
			}
			svc.Namespace = ns
			// dns environment variables
			if err := addEnv(&svc, &envs, false); err != nil {
				return err
			}
		}
	} else {
		//Inject the environment variables of dependent services according to the convention, for example, service A depends on service B,
		//Then inject {B}_HOST, {B}_PORT into the container environment variable visible to A
		var backendURLEnv apiv1.EnvVar
		for _, name := range service.Depends {
			var dependSvc *apistructs.Service
			for _, svc := range sg.Services {
				if svc.Name != name {
					continue
				}
				dependSvc = &svc
				break
			}
			// this would not happen as if this situations exist,
			// we would find out it in the stage of util.ParseServiceDependency
			if dependSvc == nil {
				return errors.Errorf("failed to find service in runtime, servicename: %s", name)
			}

			if len(dependSvc.Ports) == 0 {
				continue
			}

			// Inject {B}_HOST, {B}_PORT, etc.
			if err := addEnv(dependSvc, &envs, false); err != nil {
				return err
			}

			//Adapt the upper-level logic, if IS_ENDPOINT is assigned to true, add the BACKEND_URL environment variable
			if service.Labels["IS_ENDPOINT"] == "true" && len(dependSvc.Ports) > 0 {
				backendIP := strings.Join([]string{getServiceName(dependSvc), dependSvc.Namespace, DefaultServiceDNSSuffix}, ".")
				backendPort := dependSvc.Ports[0].Port
				backendURLEnv = apiv1.EnvVar{
					Name:  "BACKEND_URL",
					Value: backendIP + ":" + strconv.Itoa(backendPort),
				}
			}
		}

		if len(backendURLEnv.Name) > 0 {
			envs = append(envs, backendURLEnv)
			logrus.Debugf("got service BACKEND_URL, service: %s, namespace: %s, url: %s",
				serviceName, service.Namespace, backendURLEnv.Value)
		}

		// Inject the environment variables of the service itself, namely {A}_HOST, {A}_PORT, etc.
		if len(service.Ports) > 0 {
			if err := addEnv(service, &envs, false); err != nil {
				return err
			}
		}
	}

	// add K8S label
	envs = append(envs, apiv1.EnvVar{
		Name:  "IS_K8S",
		Value: "true",
	})

	// add namespace label
	envs = append(envs, apiv1.EnvVar{
		Name:  "DICE_NAMESPACE",
		Value: ns,
	})

	if service.NewHealthCheck != nil && service.NewHealthCheck.HttpHealthCheck != nil {
		envs = append(envs, apiv1.EnvVar{
			Name:  "DICE_HTTP_HEALTHCHECK_PATH",
			Value: service.NewHealthCheck.HttpHealthCheck.Path,
		})
	}
	// add POD_IP, HOST_IP
	envs = append(envs, apiv1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}, apiv1.EnvVar{
		Name: "HOST_IP",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	})

	// Add POD_NAME
	envs = append(envs, apiv1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	// Add POD_UUID
	envs = append(envs, apiv1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.uid",
			},
		},
	})

	// add SELF_URL、SELF_HOST、SELF_PORT
	selfHost := strings.Join([]string{serviceName, service.Namespace, DefaultServiceDNSSuffix}, ".")
	envs = append(envs, apiv1.EnvVar{
		Name:  "SELF_HOST",
		Value: selfHost,
	})
	for i, port := range service.Ports {
		portStr := strconv.Itoa(port.Port)
		if i == 0 {
			// TODO: we should deprecate this SELF_URL
			// TODO: we don't care what http is
			// special port env, SELF_PORT == SELF_PORT0
			envs = append(envs, apiv1.EnvVar{
				Name:  "SELF_PORT",
				Value: portStr,
			}, apiv1.EnvVar{
				Name:  "SELF_URL",
				Value: "http://" + selfHost + ":" + portStr,
			})

		}
		envs = append(envs, apiv1.EnvVar{
			Name:  "SELF_PORT" + strconv.Itoa(i),
			Value: portStr,
		})
	}

	for i, container := range containers {
		requestmem, _ := container.Resources.Requests.Memory().AsInt64()
		limitmem, _ := container.Resources.Limits.Memory().AsInt64()
		envs = append(envs,
			apiv1.EnvVar{
				Name:  "DICE_CPU_ORIGIN",
				Value: fmt.Sprintf("%f", service.Resources.Cpu)},
			apiv1.EnvVar{
				Name:  "DICE_MEM_ORIGIN",
				Value: fmt.Sprintf("%f", service.Resources.Mem)},
			apiv1.EnvVar{
				Name:  "DICE_CPU_REQUEST",
				Value: container.Resources.Requests.Cpu().AsDec().String()},
			apiv1.EnvVar{
				Name:  "DICE_MEM_REQUEST",
				Value: fmt.Sprintf("%d", requestmem/1024/1024)},
			apiv1.EnvVar{
				Name:  "DICE_CPU_LIMIT",
				Value: container.Resources.Limits.Cpu().AsDec().String()},
			apiv1.EnvVar{
				Name:  "DICE_MEM_LIMIT",
				Value: fmt.Sprintf("%d", limitmem/1024/1024)},
		)

		containers[i].Env = envs
	}
	return nil
}

func podAnnotations(service *apistructs.Service, podannotations map[string]string) {
	//Inject is turned off by default, and will not be affected by turning on inject in ns
	podannotations["sidecar.istio.io/inject"] = "false"
	// When mesh is enabled and security traffic is enabled, HTTP health checks need to be hijacked
	if service.MeshEnable != nil {
		if *service.MeshEnable {
			podannotations["sidecar.istio.io/inject"] = "true"
			if service.TrafficSecurity.Mode == "https" && service.NewHealthCheck != nil &&
				service.NewHealthCheck.HttpHealthCheck != nil {
				podannotations["sidecar.istio.io/rewriteAppHTTPProbers"] = "true"
			}
		}
	}
}

func (k *Kubernetes) newDeployment(service *apistructs.Service, sg *apistructs.ServiceGroup) (*appsv1.Deployment, error) {
	deploymentName := getDeployName(service)
	enableServiceLinks := false
	if _, ok := sg.Labels[EnableServiceLinks]; ok {
		enableServiceLinks = true
	}
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: service.Namespace,
			Labels:    make(map[string]string),
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: func(i int32) *int32 { return &i }(int32(3)),
			Replicas:             func(i int32) *int32 { return &i }(int32(service.Scale)),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deploymentName,
					Labels: make(map[string]string),
				},
				Spec: apiv1.PodSpec{
					EnableServiceLinks:    func(enable bool) *bool { return &enable }(enableServiceLinks),
					ShareProcessNamespace: func(b bool) *bool { return &b }(false),
					Tolerations:           toleration.GenTolerations(),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": service.Name},
			},
		},
	}

	if v := k.options["FORCE_BLUE_GREEN_DEPLOY"]; v != "true" &&
		(strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.DevWorkspace.String() ||
			strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.TestWorkspace.String()) {
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{Type: "Recreate"}
	}

	affinity := constraintbuilders.K8S(&sg.ScheduleInfo2, service, []constraints.PodLabelsForAffinity{
		{PodLabels: map[string]string{"app": service.Name}}}, k).Affinity
	deployment.Spec.Template.Spec.Affinity = &affinity

	// inject hosts
	deployment.Spec.Template.Spec.HostAliases = ConvertToHostAlias(service.Hosts)

	container := apiv1.Container{
		// TODO, container name e.g. redis-1528180634
		Name:  service.Name,
		Image: service.Image,
	}

	err := k.setContainerResources(*service, &container)
	if err != nil {
		errMsg := fmt.Sprintf("set container resource err: %v", err)
		logrus.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// Generate sidecars container configuration
	sidecars := k.generateSidecarContainers(service.SideCars)

	// Generate initcontainer configuration
	initcontainers := k.generateInitContainer(service.InitContainer)

	containers := []apiv1.Container{container}
	containers = append(containers, sidecars...)
	deployment.Spec.Template.Spec.Containers = containers
	if len(initcontainers) > 0 {
		deployment.Spec.Template.Spec.InitContainers = initcontainers
	}
	//for k, v := range service.Labels {
	// TODO: temporary modifications
	//if k != "HAPROXY_0_VHOST" {
	//	deployment.Metadata.Labels[k] = v
	//	deployment.Spec.Template.Metadata.Labels[k] = v
	//}
	//}

	// k8s deployment Must have app label to manage pod
	deployment.Labels["app"] = service.Name
	deployment.Spec.Template.Labels["app"] = service.Name

	setDeploymentLabels(service, deployment, sg.ID)

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	podAnnotations(service, deployment.Spec.Template.Annotations)

	// According to the current setting, there is only one user container in a pod
	if service.Cmd != "" {
		for i := range containers {
			//TODO:
			//cmds := strings.Split(service.Cmd, " ")
			cmds := []string{"sh", "-c", service.Cmd}
			containers[i].Command = cmds
		}
	}

	// Configure health check
	SetHealthCheck(&deployment.Spec.Template.Spec.Containers[0], service)

	if err := k.AddContainersEnv(containers, service, sg); err != nil {
		return nil, err
	}

	// TODO: Delete this logic
	//Mobil temporary demand:
	// Inject the secret under the "secret" namespace into the business container

	secrets, err := k.CopyErdaSecrets("secret", service.Namespace)
	if err != nil {
		logrus.Errorf("failed to copy secret: %v", err)
		return nil, err
	}
	secretvolumes := []apiv1.Volume{}
	secretvolmounts := []apiv1.VolumeMount{}
	for _, secret := range secrets {
		secretvol, volmount := k.SecretVolume(&secret)
		secretvolumes = append(secretvolumes, secretvol)
		secretvolmounts = append(secretvolmounts, volmount)
	}

	k.AddPodMountVolume(service, &deployment.Spec.Template.Spec, secretvolmounts, secretvolumes)
	k.AddSpotEmptyDir(&deployment.Spec.Template.Spec)

	logrus.Debugf("show k8s deployment, name: %s, deployment: %+v", deploymentName, deployment)
	return deployment, nil
}

func (k *Kubernetes) generateInitContainer(initcontainers map[string]diceyml.InitContainer) []apiv1.Container {
	containers := []apiv1.Container{}
	if initcontainers == nil {
		return containers
	}
	for name, initcontainer := range initcontainers {
		reqCPU := fmt.Sprintf("%.fm", k.CPUOvercommit(initcontainer.Resources.CPU)*1000)
		limitCPU := fmt.Sprintf("%.fm", initcontainer.Resources.CPU*1000)
		memory := fmt.Sprintf("%.dMi", initcontainer.Resources.Mem)

		sc := apiv1.Container{
			Name:  name,
			Image: initcontainer.Image,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(reqCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(limitCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
			},
			Command: []string{"sh", "-c", initcontainer.Cmd},
		}
		for i, dir := range initcontainer.SharedDirs {
			emptyDirVolumeName := fmt.Sprintf("%s-%d", name, i)
			dstMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.SideCar,
				ReadOnly:  false, // rw
			}
			sc.VolumeMounts = append(sc.VolumeMounts, dstMount)
		}
		containers = append(containers, sc)
	}

	return containers
}

func (k *Kubernetes) generateSidecarContainers(sidecars map[string]*diceyml.SideCar) []apiv1.Container {
	var containers []apiv1.Container

	if len(sidecars) == 0 {
		return nil
	}

	for name, sidecar := range sidecars {
		reqCPU := fmt.Sprintf("%.fm", k.CPUOvercommit(sidecar.Resources.CPU)*1000)
		limitCPU := fmt.Sprintf("%.fm", sidecar.Resources.CPU*1000)
		memory := fmt.Sprintf("%.dMi", sidecar.Resources.Mem)

		sc := apiv1.Container{
			Name:  strutil.Concat(sidecarNamePrefix, name),
			Image: sidecar.Image,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(reqCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(limitCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
			},
		}

		//Do not insert platform environment variables (DICE_*) to prevent the instance list from being collected
		for k, v := range sidecar.Envs {
			sc.Env = append(sc.Env, apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		// Sidecar and business container share directory
		for _, dir := range sidecar.SharedDirs {
			emptyDirVolumeName := strutil.Concat(name, shardDirSuffix)
			dstMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.SideCar,
				ReadOnly:  false, // rw
			}
			sc.VolumeMounts = append(sc.VolumeMounts, dstMount)
		}
		containers = append(containers, sc)
	}

	return containers
}

func makeEnvVariableName(str string) string {
	return strings.ToUpper(strings.Replace(str, "-", "_", -1))
}

// AddPodMountVolume Add pod volume configuration
func (k *Kubernetes) AddPodMountVolume(service *apistructs.Service, podSpec *apiv1.PodSpec,
	secretvolmounts []apiv1.VolumeMount, secretvolumes []apiv1.Volume) error {

	podSpec.Volumes = make([]apiv1.Volume, 0)

	//Pay attention to the settings mentioned above, there is only one container in a pod
	podSpec.Containers[0].VolumeMounts = make([]apiv1.VolumeMount, 0)

	// get cluster info
	clusterInfo, err := k.ClusterInfo.Get()
	if err != nil {
		return errors.Errorf("failed to get cluster info, clusterName: %s, (%v)", k.clusterName, err)
	}

	// hostPath type
	for i, bind := range service.Binds {
		if bind.HostPath == "" || bind.ContainerPath == "" {
			continue
		}
		//Name formation '[a-z0-9]([-a-z0-9]*[a-z0-9])?'
		name := "volume" + "-bind-" + strconv.Itoa(i)

		hostPath, err := ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
		if err != nil {
			return err
		}
		if !strutil.HasPrefixes(hostPath, "/") {
			pvcName := strings.Replace(hostPath, "_", "-", -1)
			sc := "dice-local-volume"
			if err := k.pvc.CreateIfNotExists(&apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s", service.Name, pvcName),
					Namespace: service.Namespace,
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					StorageClassName: &sc,
				},
			}); err != nil {
				return err
			}

			podSpec.Volumes = append(podSpec.Volumes,
				apiv1.Volume{
					Name: name,
					VolumeSource: apiv1.VolumeSource{
						PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
							ClaimName: fmt.Sprintf("%s-%s", service.Name, pvcName),
						},
					},
				})

			podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts,
				apiv1.VolumeMount{
					Name:      name,
					MountPath: bind.ContainerPath,
					ReadOnly:  bind.ReadOnly,
				})
			continue
		}

		podSpec.Volumes = append(podSpec.Volumes,
			apiv1.Volume{
				Name: name,
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: hostPath,
					},
				},
			})

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts,
			apiv1.VolumeMount{
				Name:      name,
				MountPath: bind.ContainerPath,
				ReadOnly:  bind.ReadOnly,
			})
	}

	// Configure the business container sidecar shared directory
	for name, sidecar := range service.SideCars {
		for _, dir := range sidecar.SharedDirs {
			emptyDirVolumeName := strutil.Concat(name, shardDirSuffix)

			srcMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.Main,
				ReadOnly:  false, // rw
			}
			// Business master container
			podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, srcMount)

			podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
				Name: emptyDirVolumeName,
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			})
		}
	}
	if service.InitContainer != nil {
		for name, initc := range service.InitContainer {
			for i, dir := range initc.SharedDirs {
				name := fmt.Sprintf("%s-%d", name, i)
				srcMount := apiv1.VolumeMount{
					Name:      name,
					MountPath: dir.Main,
					ReadOnly:  false,
				}
				podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, srcMount)
				podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
					Name: name,
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				})
			}
		}
	}

	podSpec.Volumes = append(podSpec.Volumes, secretvolumes...)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, secretvolmounts...)

	return nil
}
func (k *Kubernetes) AddSpotEmptyDir(podSpec *apiv1.PodSpec) {
	podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
		Name:         "spot-emptydir",
		VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{}},
	})

	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, apiv1.VolumeMount{
		Name:      "spot-emptydir",
		MountPath: "/tmp",
	})
}

func getDeployName(service *apistructs.Service) string {
	if service.ProjectServiceName != "" {
		return service.ProjectServiceName
	}
	return service.Name
}

func setDeploymentLabels(service *apistructs.Service, deployment *appsv1.Deployment, sgID string) {
	if service.ProjectServiceName != "" {
		deployment.Spec.Selector.MatchLabels[LabelServiceGroupID] = sgID
		deployment.Spec.Template.Labels[LabelServiceGroupID] = sgID
		deployment.Labels[LabelServiceGroupID] = sgID
	}
}

func ConvertToHostAlias(hosts []string) []apiv1.HostAlias {
	var r []apiv1.HostAlias
	for _, host := range hosts {
		splitRes := strings.Fields(host)
		if len(splitRes) < 2 {
			continue
		}
		r = append(r, apiv1.HostAlias{
			IP:        splitRes[0],
			Hostnames: splitRes[1:],
		})
	}
	return r
}

func (k *Kubernetes) scaleDeployment(sg *apistructs.ServiceGroup) error {
	// only support scale the first one service
	ns := sg.ProjectNamespace
	if ns == "" {
		ns = MakeNamespace(sg)
	}

	scalingService := sg.Services[0]
	deploymentName := getDeployName(&scalingService)
	deploy, err := k.getDeployment(ns, deploymentName)
	if err != nil {
		getErr := fmt.Errorf("failed to get the deployment, err is: %s", err.Error())
		return getErr
	}

	deploy.Spec.Replicas = func(i int32) *int32 { return &i }(int32(scalingService.Scale))

	// only support one container on Erda currently
	container := deploy.Spec.Template.Spec.Containers[0]

	err = k.setContainerResources(scalingService, &container)
	if err != nil {
		setContainerErr := fmt.Errorf("failed to set container resource, err is: %s", err.Error())
		return setContainerErr
	}

	deploy.Spec.Template.Spec.Containers[0] = container
	err = k.deploy.Put(deploy)
	if err != nil {
		updateErr := fmt.Errorf("failed to update the deployment, err is: %s", err.Error())
		return updateErr
	}
	return nil
}

func (k *Kubernetes) setContainerResources(service apistructs.Service, container *apiv1.Container) error {
	if service.Resources.Cpu < MIN_CPU_SIZE {
		return errors.Errorf("invalid cpu, value: %v, (which is lower than min cpu(%v))",
			service.Resources.Cpu, MIN_CPU_SIZE)
	}

	//Set the over-score ratio according to the environment
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
	if cpuSubscribeRatio < 1.0 {
		cpuSubscribeRatio = 1.0
	}
	if memSubscribeRatio < 1.0 {
		memSubscribeRatio = 1.0
	}
	requestCPU := "10m"
	if service.Resources.Cpu*1000/cpuSubscribeRatio > 10.0 {
		requestCPU = fmt.Sprintf("%dm", int(service.Resources.Cpu*1000/cpuSubscribeRatio))
	}

	requestMem := fmt.Sprintf("%dMi", int(service.Resources.Mem/memSubscribeRatio))

	cpu := fmt.Sprintf("%dm", int(service.Resources.Cpu*1000))
	memory := fmt.Sprintf("%dMi", int(service.Resources.Mem))

	container.Resources = apiv1.ResourceRequirements{
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse(requestCPU),
			apiv1.ResourceMemory: resource.MustParse(requestMem),
		},
		Limits: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse(cpu),
			apiv1.ResourceMemory: resource.MustParse(memory),
		},
	}

	return nil
}
