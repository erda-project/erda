package elasticsearch

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ElasticsearchAndSecret struct {
	Elasticsearch
	corev1.Secret
}

type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ElasticsearchSpec `json:"spec"`
}

type ElasticsearchSpec struct {
	Http     HttpSettings       `json:"http,omitempty"`
	Version  string             `json:"version,omitempty"`
	Image    string             `json:"image,omitempty"`
	NodeSets []NodeSetsSettings `json:"nodeSets,omitempty"`
}

// HttpSettings
type HttpSettings struct {
	Tls TlsSettings `json:"tls,omitempty"`
}

type TlsSettings struct {
	SelfSignedCertificate SelfSignedCertificateSettings `json:"selfSignedCertificate,omitempty"`
}

type SelfSignedCertificateSettings struct {
	Disabled bool `json:"disabled,omitempty"`
}

// NodeSetsSettings
type NodeSetsSettings struct {
	Name                 string                `json:"name,omitempty"`
	Count                int                   `json:"count,omitempty"`
	Config               map[string]string     `json:"config,omitempty"`
	PodTemplate          PodTemplateSettings   `json:"podTemplate,omitempty"`
	VolumeClaimTemplates []VolumeClaimSettings `json:"volumeClaimTemplates,omitempty"`
}

type PodTemplateSettings struct {
	Spec PodSpecSettings `json:"spec,omitempty"`
}

type PodSpecSettings struct {
	Affinity   *corev1.Affinity     `json:"affinity,omitempty"`
	Containers []ContainersSettings `json:"containers,omitempty"`
}

type ContainersSettings struct {
	Name      string                      `json:"name,omitempty"`
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// VolumeClaimSettings for ElasticsearchSpec NodeSetsSettings
type VolumeClaimSettings struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VolumeClaimSpecSettings `json:"spec,omitempty"`
}

type VolumeClaimSpecSettings struct {
	AccessModes      []string                    `json:"accessModes,omitempty"`
	Resources        corev1.ResourceRequirements `json:"resources,omitempty"`
	StorageClassName string                      `json:"storageClassName,omitempty"`
}
