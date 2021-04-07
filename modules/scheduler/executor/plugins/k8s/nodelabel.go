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

package k8s

import (
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func (m *Kubernetes) IPToHostname(ip string) string {
	nodelist, err := m.nodeLabel.List()
	if err != nil {
		return ""
	}
	for _, node := range nodelist.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP && addr.Address == ip {
				return node.Name
			}
		}
	}
	return ""
}

// SetNodeLabels set the labels of k8s node
func (m *Kubernetes) SetNodeLabels(_ executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	// contents in 'hosts' maybe hostname or internalIP, it should be unified into hostname
	nodelist, err := m.nodeLabel.List()
	if err != nil {
		logrus.Errorf("failed to list nodes: %v", err)
		return err
	}
	updatedHosts := []string{}
	for _, host := range hosts {
		for _, node := range nodelist.Items {
			add := false
			for _, addr := range node.Status.Addresses {

				if addr.Address == host {
					add = true
					break
				}
			}
			if add {
				updatedHosts = append(updatedHosts, node.Name)
			}
		}
	}

	for _, host := range updatedHosts {
		prefixedLabels := map[string]*string{}
		orig, err := m.nodeLabel.Get(host)
		if err != nil {
			return err
		}

		// 1. unset all 'dice/' labels
		for k := range orig {
			if !strutil.HasPrefixes(k, labelconfig.K8SLabelPrefix) {
				continue
			}
			prefixedLabels[k] = nil
		}

		// 2. set labels in param 'labels'
		for k := range labels {
			v := labels[k]
			prefixedKey := k
			if !strutil.HasPrefixes(prefixedKey, labelconfig.K8SLabelPrefix) {
				prefixedKey = strutil.Concat(labelconfig.K8SLabelPrefix, k)
			}
			prefixedLabels[prefixedKey] = &v
		}

		// 3. set them
		if err := m.nodeLabel.Set(prefixedLabels, host); err != nil {
			return err
		}
	}
	return nil
}
