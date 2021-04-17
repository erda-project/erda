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

package netdatavolume

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/strutil"
)

func (p *netDataVolumeProvisioner) Delete(ctx context.Context, pv *v1.PersistentVolume) error {
	logrus.Infof("Start deleting volume: %s/%s", pv.Namespace, pv.Name)
	volPathInContainer := strutil.JoinPath("/hostfs", pv.Spec.PersistentVolumeSource.Local.Path)
	if err := os.RemoveAll(volPathInContainer); err != nil {
		logrus.Errorf("Failed to remove path: %v", volPathInContainer)
	}
	return nil
}
