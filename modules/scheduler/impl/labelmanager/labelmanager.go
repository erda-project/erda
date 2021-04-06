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

package labelmanager

// LabelManager interface of LabelManagerImpl
type LabelManager interface {
	// List available labels.
	// map key: label name
	// map value: Is this label a prefix?
	//
	// Although different types of
	// clusters (dcos&k8s) have different names on the compute
	// nodes, the exposed labels are uniform names.
	List() map[string]bool

	// Not yet implemented
	SetNodeLabel(cluster Cluster, hosts []string, tags map[string]string) error

	// Not yet implemented
	// GetNodeLabel() map[string]string
}

// LabelManagerImpl implementation of LabelManager
type LabelManagerImpl struct{}

type Cluster struct {
	// e.g. terminus-y
	ClusterName string
	// One of the following options:
	// ServiceKindMarathon
	// ServiceKindK8S
	// JobKindMetronome
	// JobKindK8S
	ClusterType string
	// This can be ignored for k8s cluster
	SoldierURL string
}

// NewLabelManager create LabelManager
func NewLabelManager() LabelManager {
	return &LabelManagerImpl{}
}
