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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CanalSpec defines the desired state of Canal
type CanalSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:default=v1.1.5
	//+optional
	Version string `json:"version,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=9
	//+kubebuilder:default=1
	//+optional
	Replicas int `json:"replicas,omitempty"`

	//+optional
	CanalOptions map[string]string `json:"canalOptions,omitempty"`
	//canal.admin.manager 127.0.0.1:8089
	//canal.admin.port    11110
	//canal.admin.user    admin
	//canal.admin.passwd  admin

	//+optional
	AdminOptions map[string]string `json:"adminOptions,omitempty"`
	//spring.datasource.address  127.0.0.1:3306
	//spring.datasource.database canal_manager
	//spring.datasource.username canal
	//spring.datasource.password canal
	//canal.adminUser            admin
	//canal.adminPasswd          admin

	//+optional
	JavaOptions string `json:"javaOptions,omitempty"`

	//+optional
	Image string `json:"image,omitempty"`
	//+kubebuilder:default=IfNotPresent
	//+optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	// +optional
	AdminResources corev1.ResourceRequirements `json:"adminResources,omitempty" protobuf:"bytes,8,opt,name=adminResources"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
}

const (
	Red    = "red"
	Yellow = "yellow"
	Green  = "green"
)

// CanalStatus defines the observed state of Canal
type CanalStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+optional
	Color string `json:"color,omitempty"`

	//+optional
	AdminInitialized bool `json:"adminInitialized,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Color",type=string,JSONPath=`.status.color`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Canal is the Schema for the canals API
type Canal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanalSpec   `json:"spec,omitempty"`
	Status CanalStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CanalList contains a list of Canal
type CanalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Canal `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Canal{}, &CanalList{})
}

func (r *Canal) NewLabels() map[string]string {
	return map[string]string{
		"addon": "canal",
		"group": r.Name,
	}
}

const HeadlessSuffix = "x"

func (r *Canal) BuildName(suffix string) string {
	return r.Name + "-" + suffix
}
