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

package legacy

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RedisFailover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisFailoverSpec   `json:"spec"`
	Status            RedisFailoverStatus `json:"status,omitempty"`
}

// RedisFailoverSpec represents a Redis failover spec
type RedisFailoverSpec struct {
	// Redis defines its failover settings
	Redis RedisSettings `json:"redis,omitempty"`

	// Sentinel defines its failover settings
	Sentinel SentinelSettings `json:"sentinel,omitempty"`

	// HardAntiAffinity defines if the PodAntiAffinity on the deployments and
	// statefulsets has to be hard (it's soft by default)
	HardAntiAffinity bool `json:"hardAntiAffinity,omitempty"`

	// NodeAffinity defines the rules for scheduling the Redis and Sentinel
	// nodes
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	// SecurityContext defines which user and group the Sentinel and Redis containers run as
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	//Tolerations provides a way to schedule Pods on Tainted Nodes
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Envs                  map[string]string      `json:"envs,omitempty"`
	Replicas              int32                  `json:"replicas,omitempty"`
	Resources             RedisFailoverResources `json:"resources,omitempty"`
	Exporter              bool                   `json:"exporter,omitempty"`
	ExporterImage         string                 `json:"exporterImage,omitempty"`
	ExporterVersion       string                 `json:"exporterVersion,omitempty"`
	DisableExporterProbes bool                   `json:"disableExporterProbes,omitempty"`
	Image                 string                 `json:"image,omitempty"`
	Version               string                 `json:"version,omitempty"`
	CustomConfig          []string               `json:"customConfig,omitempty"`
	Command               []string               `json:"command,omitempty"`
	ShutdownConfigMap     string                 `json:"shutdownConfigMap,omitempty"`
	Storage               RedisStorage           `json:"storage,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Envs         map[string]string      `json:"envs,omitempty"`
	Replicas     int32                  `json:"replicas,omitempty"`
	Resources    RedisFailoverResources `json:"resources,omitempty"`
	CustomConfig []string               `json:"customConfig,omitempty"`
	Command      []string               `json:"command,omitempty"`
}

// RedisFailoverResources sets the limits and requests for a container
type RedisFailoverResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// RedisFailoverStatus has the status of the cluster
type RedisFailoverStatus struct {
	Phase      Phase       `json:"phase"`
	Conditions []Condition `json:"conditions"`
	Master     string      `json:"master"`
}

// Phase of the RF status
type Phase string

// Condition saves the state information of the redisfailover
type Condition struct {
	Type           ConditionType `json:"type"`
	Reason         string        `json:"reason"`
	TransitionTime string        `json:"transitionTime"`
}

// ConditionType defines the condition that the RF can have
type ConditionType string
