package apistructs

import corev1 "k8s.io/api/core/v1"

type FlinkKind string

const (
	FlinkJob     FlinkKind = "FlinkJob"
	FlinkSession FlinkKind = "FlinkSession"
)

type BigdataResource struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Replica int32  `json:"replica"`
}

type BigdataMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type BigdataSpec struct {
	Image      string            `json:"image,omitempty"`
	Resource   string            `json:"resource,omitempty"`
	Class      string            `json:"class,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Envs       []corev1.EnvVar   `json:"envs,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	FlinkConf  *FlinkConf        `json:"flinkConf,omitempty"`
	SparkConf  *SparkConf        `json:"sparkConf,omitempty"`
}

type BigdataConf struct {
	BigdataMetadata `json:"metadata"`
	Spec            BigdataSpec `json:"spec"`
}

type FlinkConf struct {
	Kind                FlinkKind       `json:"kind"`
	Parallelism         int32           `json:"parallelism"`
	JobManagerResource  BigdataResource `json:"jobManagerResource"`
	TaskManagerResource BigdataResource `json:"taskManagerResource"`
}

type SparkConf struct {
	Type             string          `json:"type"` // support Java, Python, Scala and R
	Kind             string          `json:"kind"` // support Client and Cluster
	PythonVersion    *string         `json:"pythonVersion,omitempty"`
	DriverResource   BigdataResource `json:"driverResource"`
	ExecutorResource BigdataResource `json:"executorResource"`
}
