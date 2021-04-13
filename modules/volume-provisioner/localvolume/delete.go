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

package localvolume

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/strutil"
)

func (p *localVolumeProvisioner) Delete(ctx context.Context, pv *v1.PersistentVolume) error {
	logrus.Infof("Start deleting volume: namespace: %s, pvname: %v", pv.Namespace, pv.Name)

	nodeListOption, err := genListOptionFromNodeAffinity(pv.Spec.NodeAffinity)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return p.cmdExecutor.OnNodesPods(fmt.Sprintf("rm -rf %s || true",
		strutil.JoinPath("/hostfs", pv.Spec.PersistentVolumeSource.Local.Path)),
		nodeListOption, metav1.ListOptions{LabelSelector: "app=volume-provisioner"})
}

func genListOptionFromNodeAffinity(affinity *v1.VolumeNodeAffinity) (metav1.ListOptions, error) {
	for _, t := range affinity.Required.NodeSelectorTerms {
		for _, expr := range t.MatchExpressions {
			if expr.Key == "kubernetes.io/hostname" &&
				expr.Operator == v1.NodeSelectorOpIn &&
				len(expr.Values) == 1 {
				return metav1.ListOptions{
					LabelSelector: fmt.Sprintf("kubernetes.io/hostname=%s", expr.Values[0]),
				}, nil
			}
		}
	}
	return metav1.ListOptions{}, fmt.Errorf("Failed to generate ListOption from VolumeNodeAffinity: %v", affinity)
}
