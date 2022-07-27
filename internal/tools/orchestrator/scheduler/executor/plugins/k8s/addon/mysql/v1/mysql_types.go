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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MysqlSpec defines the desired state of Mysql
type MysqlSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=v5.7;v8.0
	//+kubebuilder:default=v5.7
	//+optional
	Version string `json:"version,omitempty"`

	//+kubebuilder:validation:Enum=Classic;Single;Multi
	//+kubebuilder:default=Classic
	//+optional
	PrimaryMode string `json:"primaryMode,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=9
	//+kubebuilder:default=1
	//+optional
	Primaries int `json:"primaries,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=9
	//+optional
	Replicas *int `json:"replicas,omitempty"`

	//+optional
	PrimaryId *int `json:"primaryId,omitempty"`

	//+optional
	AutoSwitch *bool `json:"autoSwitch,omitempty"`

	//+kubebuilder:default=root
	//+optional
	LocalUsername string `json:"localUsername,omitempty"`
	//+optional
	LocalPassword string `json:"localPassword,omitempty"`
	//+kubebuilder:default=repl
	//+optional
	ReplicaUsername string `json:"replicaUsername,omitempty"`
	//+optional
	ReplicaPassword string `json:"replicaPassword,omitempty"`

	//+kubebuilder:default=standard
	//+optional
	StorageClassName string `json:"storageClassName,omitempty"`
	//+kubebuilder:default="10Gi"
	//+optional
	StorageSize resource.Quantity `json:"storageSize,omitempty"`

	//+optional
	Image string `json:"image,omitempty"`
	//+kubebuilder:default=IfNotPresent
	//+optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default=3306
	//+optional
	Port int `json:"port,omitempty"`
	//+kubebuilder:default=33080
	//+optional
	MyletPort int `json:"myletPort,omitempty"`

	//+kubebuilder:default=/mydir
	//+optional
	Mydir string `json:"mydir,omitempty"`
	//+optional
	MyctlAddr string `json:"myctlAddr,omitempty"`
	//+optional
	HeadlessHost string `json:"headlessHost,omitempty"`
	//+optional
	ShortHeadlessHost string `json:"shortHeadlessHost,omitempty"`

	//+optional
	Solos []MysqlSoloSpec `json:"solos,omitempty"`

	//+optional
	//+kubebuilder:default=false
	EnableExporter bool `json:"enableExporter,omitempty"`
	//+kubebuilder:default=9104
	ExporterPort int `json:"exporterPort,omitempty"`
	//+optional
	ExporterFlags []string `json:"exporterFlags,omitempty"`
	//+optional
	ExporterImage string `json:"exporterImage,omitempty"`
	//+kubebuilder:default=exporter
	//+optional
	ExporterUsername string `json:"exporterUsername,omitempty"`
	//+optional
	ExporterPassword string `json:"exporterPassword,omitempty"`

	//+kubebuilder:default=33061
	//+optional
	GroupPort int `json:"groupPort,omitempty"`
	//+optional
	GroupName string `json:"groupName,omitempty"`

	//+optional
	Labels map[string]string `json:"labels,omitempty"`
	//+optional
	Annotations map[string]string `json:"annotations,omitempty"`
	//+optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	//+optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+optional
	Version MysqlVersion `json:"version,omitempty"`
	//+optional
	Solos []MysqlSolo `json:"solos,omitempty"`
	//+optional
	MysqlSoloStatus `json:",inline,omitempty"`

	//+optional
	WriteId *int `json:"writeId,omitempty"`
	//+optional
	ReadId *int `json:"readId,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Primaries",type=integer,JSONPath=`.spec.primaries`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Color",type=string,JSONPath=`.status.color`
//+kubebuilder:printcolumn:name="WriteId",type=integer,JSONPath=`.status.writeId`
//+kubebuilder:printcolumn:name="ReadId",type=integer,JSONPath=`.status.readId`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Mysql is the Schema for the mysqls API
type Mysql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlSpec   `json:"spec,omitempty"`
	Status MysqlStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlList contains a list of Mysql
type MysqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysql `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mysql{}, &MysqlList{})
}

func (r *Mysql) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

func (r *Mysql) NewLabels() map[string]string {
	return map[string]string{
		"addon": "mysql",
		"group": r.Name,
	}
}

const HeadlessSuffix = "x"

func (r *Mysql) BuildName(suffix string) string {
	return r.Name + "-" + suffix
}

func (spec *MysqlSpec) Size() int {
	return spec.Primaries + *spec.Replicas
}
