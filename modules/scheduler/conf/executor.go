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

// Org organization, Corresponding tenant concept
type Org struct {
	Name       string            `json:"name,omitempty"`
	Workspaces []Workspace       `json:"workspaces,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

// Environment
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
