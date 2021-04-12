// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package k8sflink

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/erda-project/erda/apistructs"
	flinkv1beta1 "github.com/erda-project/erda/pkg/clientgo/apis/flinkoperator/v1beta1"
)

const (
	AliyunPullSecret   = "aliyun-registry"
	FlinkIngressPrefix = "flinkcluster"
)

func getInt32Points(numeric int32) *int32 {
	return &numeric
}
func getStringPoints(s string) *string {
	return &s
}

func composeResources(res apistructs.BigdataResource) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(res.CPU),
			corev1.ResourceMemory: resource.MustParse(res.Memory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(res.CPU),
			corev1.ResourceMemory: resource.MustParse(res.Memory),
		},
	}
}

func composeEnvs(envs map[string]string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}

	for k, v := range envs {
		envVars = append(envVars, corev1.EnvVar{
			Name:      k,
			Value:     v,
			ValueFrom: nil,
		})
	}

	return envVars
}

func ComposeFlinkCluster(data apistructs.BigdataConf, hostURL string) *flinkv1beta1.FlinkCluster {

	flinkCluster := flinkv1beta1.FlinkCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FlinkCluster",
			APIVersion: "flinkoperator.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        data.Name,
			Namespace:   data.Namespace,
			Labels:      nil,
			Annotations: nil,
		},
		Spec: flinkv1beta1.FlinkClusterSpec{
			Image: flinkv1beta1.ImageSpec{
				Name:       data.Spec.Image,
				PullPolicy: corev1.PullAlways,
				PullSecrets: []corev1.LocalObjectReference{
					{
						Name: AliyunPullSecret,
					},
				},
			},
			JobManager: flinkv1beta1.JobManagerSpec{
				Ingress: &flinkv1beta1.JobManagerIngressSpec{
					HostFormat: getStringPoints(hostURL),
				},
				Replicas:       getInt32Points(data.Spec.FlinkConf.JobManagerResource.Replica),
				Resources:      composeResources(data.Spec.FlinkConf.JobManagerResource),
				Volumes:        nil,
				VolumeMounts:   nil,
				InitContainers: nil,
				NodeSelector:   nil,
				Tolerations:    nil,
				Sidecars:       nil,
				PodAnnotations: nil,
			},
			TaskManager: flinkv1beta1.TaskManagerSpec{
				Replicas:       data.Spec.FlinkConf.TaskManagerResource.Replica,
				Resources:      composeResources(data.Spec.FlinkConf.TaskManagerResource),
				Volumes:        nil,
				VolumeMounts:   nil,
				InitContainers: nil,
				NodeSelector:   nil,
				Tolerations:    nil,
				Sidecars:       nil,
				PodAnnotations: nil,
			},
			EnvVars:         data.Spec.Envs,
			FlinkProperties: data.Spec.Properties,
			LogConfig:       composeLogConfig(),
		},
	}

	if data.Spec.FlinkConf.Kind == apistructs.FlinkJob {
		flinkCluster.Spec.Job = composeFlinkJob(data)
	}

	return &flinkCluster
}

func composeFlinkJob(data apistructs.BigdataConf) *flinkv1beta1.JobSpec {
	return &flinkv1beta1.JobSpec{
		JarFile:           data.Spec.Resource,
		ClassName:         &data.Spec.Class,
		Args:              data.Spec.Args,
		Parallelism:       getInt32Points(data.Spec.FlinkConf.Parallelism),
		NoLoggingToStdout: nil,
		Volumes:           nil,
		VolumeMounts:      nil,
		InitContainers:    nil,
		RestartPolicy:     getJobRestartPolicy(flinkv1beta1.JobRestartPolicyFromSavepointOnFailure),
		CleanupPolicy: &flinkv1beta1.CleanupPolicy{
			AfterJobSucceeds:  flinkv1beta1.CleanupActionDeleteCluster,
			AfterJobFails:     flinkv1beta1.CleanupActionKeepCluster,
			AfterJobCancelled: flinkv1beta1.CleanupActionDeleteTaskManager,
		},
		CancelRequested: nil,
		PodAnnotations:  nil,
		Resources:       corev1.ResourceRequirements{},
	}
}

func getJobRestartPolicy(restartPolicy flinkv1beta1.JobRestartPolicy) *flinkv1beta1.JobRestartPolicy {
	return &restartPolicy
}

func composeStatusDesc(status flinkv1beta1.FlinkClusterStatus) apistructs.StatusDesc {
	statusDesc := apistructs.StatusDesc{}
	switch status.State {
	case flinkv1beta1.ClusterStateCreating,
		flinkv1beta1.ClusterStateReconciling,
		flinkv1beta1.ClusterStateUpdating,
		flinkv1beta1.ClusterStateRunning:
		statusDesc.Status = apistructs.StatusRunning
	case flinkv1beta1.ClusterStateStopping,
		flinkv1beta1.ClusterStatePartiallyStopped,
		flinkv1beta1.ClusterStateStopped:
		statusDesc.Status = apistructs.StatusStopped
	}
	if status.Components.Job == nil {
		return statusDesc
	}
	switch status.Components.Job.State {
	case flinkv1beta1.JobStateSucceeded:
		statusDesc.Status = apistructs.StatusStoppedOnOK
	case flinkv1beta1.JobStateFailed:
		statusDesc.Status = apistructs.StatusStoppedOnFailed
	case flinkv1beta1.JobStateCancelled:
		statusDesc.Status = apistructs.StatusStoppedByKilled
	}
	return statusDesc
}

func getBoolPoint(bl bool) *bool {
	return &bl
}

func composeOwnerReferences(versionGroup, kind, name string, uid types.UID) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         versionGroup,
		Kind:               kind,
		Name:               name,
		UID:                uid,
		Controller:         getBoolPoint(true),
		BlockOwnerDeletion: getBoolPoint(true),
	}
}

func composeLogConfig() map[string]string {
	logConfig := map[string]string{
		"log4j-console.properties": `rootLogger.level = INFO
      rootLogger.appenderRef.file.ref = LogFile
      rootLogger.appenderRef.console.ref = LogConsole
      appender.file.name = LogFile
      appender.file.type = File
      appender.file.append = false
      appender.file.fileName = ${sys:log.file}
      appender.file.layout.type = PatternLayout
      appender.file.layout.pattern = %d{yyyy-MM-dd HH:mm:ss,SSS} %-5p %-60c %x - %m%n
      appender.console.name = LogConsole
      appender.console.type = CONSOLE
      appender.console.layout.type = PatternLayout
      appender.console.layout.pattern = %d{yyyy-MM-dd HH:mm:ss,SSS} %-5p %-60c %x - %m%n
      logger.akka.name = akka
      logger.akka.level = INFO
      logger.kafka.name= org.apache.kafka
      logger.kafka.level = INFO
      logger.hadoop.name = org.apache.hadoop
      logger.hadoop.level = INFO
      logger.zookeeper.name = org.apache.zookeeper
      logger.zookeeper.level = INFO
      logger.netty.name = org.apache.flink.shaded.akka.org.jboss.netty.channel.DefaultChannelPipeline
      logger.netty.level = OFF`,
		"logback-console.xml": `<configuration>
        <appender name="console" class="ch.qos.logback.core.ConsoleAppender">
          <encoder>
            <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{60} %X{sourceThread} - %msg%n</pattern>
          </encoder>
        </appender>
        <appender name="file" class="ch.qos.logback.core.FileAppender">
          <file>${log.file}</file>
          <append>false</append>
          <encoder>
            <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{60} %X{sourceThread} - %msg%n</pattern>
          </encoder>
        </appender>
        <root level="INFO">
          <appender-ref ref="console"/>
          <appender-ref ref="file"/>
        </root>
        <logger name="akka" level="INFO" />
        <logger name="org.apache.kafka" level="INFO" />
        <logger name="org.apache.hadoop" level="INFO" />
        <logger name="org.apache.zookeeper" level="INFO" />
        <logger name="org.apache.flink.shaded.akka.org.jboss.netty.channel.DefaultChannelPipeline" level="INFO" />
      </configuration>`,
	}
	return logConfig
}
