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

// Package volumeprovisioner is a persistent volume claim provisioner for kubernetes.
package volumeprovisioner

import (
	"context"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"

	"github.com/erda-project/erda/modules/volume-provisioner/localvolume"
	"github.com/erda-project/erda/modules/volume-provisioner/netdatavolume"
)

func initLocalVolumeProvisioner(config *config, csConfig *rest.Config, client kubernetes.Interface, version *version.Info) {
	var pc *controller.ProvisionController

	logrus.Infof("Creating localvolumeProvisioner...")

	lvpConfig := &localvolume.Config{
		ModeEdge:   config.ModeEdge,
		MatchLabel: config.LocalMatchLabel,
		NodeName:   config.NodeName,
		Namespace:  config.ProvisionerNamespace,
	}

	lvp := localvolume.NewLocalVolumeProvisioner(lvpConfig, csConfig, client)

	if config.ModeEdge {
		pc = controller.NewProvisionController(client, config.LocalProvisionerName, lvp, version.GitVersion,
			controller.LeaderElection(false))
	} else {
		pc = controller.NewProvisionController(client, config.LocalProvisionerName, lvp, version.GitVersion)
	}

	go pc.Run(context.Background())
}

func initNetDataVolumeProvisioner(config *config, csConfig *rest.Config, client kubernetes.Interface, version *version.Info) {
	logrus.Infof("Creating netdatavolumeProvisioner...")

	nvp := netdatavolume.NewNetDataVolumeProvisioner(csConfig, client)
	pc := controller.NewProvisionController(client, config.NetProvisionerName, nvp, version.GitVersion)

	go pc.Run(context.Background())
}

func initialize(config *config, ctx context.Context) error {

	csConfig, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("Failed to create config: %v", err)
		return err
	}
	cs, err := kubernetes.NewForConfig(csConfig)
	if err != nil {
		logrus.Errorf("Failed to create client: %v", err)
		return err
	}
	serverVersion, err := cs.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Failed to get server version: %v", err)
		return err
	}

	if config.ModeEdge {
		logrus.Infof("Edge mode, create localvolumeProvisioner only")
		initLocalVolumeProvisioner(config, csConfig, cs, serverVersion)
	} else {
		initLocalVolumeProvisioner(config, csConfig, cs, serverVersion)
		initNetDataVolumeProvisioner(config, csConfig, cs, serverVersion)
	}

	select {
	case <-ctx.Done():
	}
	return nil
}
