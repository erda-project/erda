package conf

import "sync"

const (
	CLUSTERS_CONFIG_PATH = "/dice/scheduler/configs/cluster/"
)

type ExecutorConfig struct {
	Kind        string            `json:"kind,omitempty"`
	Name        string            `json:"name,omitempty"`
	ClusterName string            `json:"clusterName,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	OptionsPlus *OptPlus          `json:"optionsPlus,omitempty"`
}

type OptPlus struct {
	Orgs []Org `json:"orgs,omitempty"`
}

// Org 结构体，对应租户概念
type Org struct {
	Name       string            `json:"name,omitempty"`
	Workspaces []Workspace       `json:"workspaces,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

// 发布的环境，name 从 dev, test, staging, prod 中选取
type Workspace struct {
	Name    string            `json:"name,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

var confStore ConfStore

func GetConfStore() *ConfStore {
	return &confStore
}

type ConfStore struct {
	ExecutorStore sync.Map
}
