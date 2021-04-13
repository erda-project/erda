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

const (
	// name of local volume provisioner
	localvolumeProvisioner = "dice/local-volume"
	// name of netdata volume provisioner
	netdataVolumeProvisioner = "dice/netdata-volume"
	// namespace of volume provisioner
	namespace = "default"
)

func initLocalVolumeProvisioner(config *rest.Config, client kubernetes.Interface, version *version.Info) {
	logrus.Infof("Creating localvolumeProvisioner...")
	lvp := localvolume.NewLocalVolumeProvisioner(config, client, namespace)
	pc := controller.NewProvisionController(client, localvolumeProvisioner, lvp, version.GitVersion)
	go pc.Run(context.Background())
}

func initNetdataVolumeProvisioner(config *rest.Config, client kubernetes.Interface, version *version.Info) {
	logrus.Infof("Creating netdatavolumeProvisioner...")
	nvp := netdatavolume.NewNetdataVolumeProvisioner(config, client, namespace)
	pc := controller.NewProvisionController(client, netdataVolumeProvisioner, nvp, version.GitVersion)
	go pc.Run(context.Background())
}

func initialize() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("Failed to create config: %v", err)
		return err
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("Failed to create client: %v", err)
		return err
	}
	serverVersion, err := cs.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Failed to get server version: %v", err)
		return err
	}

	initLocalVolumeProvisioner(config, cs, serverVersion)
	initNetdataVolumeProvisioner(config, cs, serverVersion)

	ch := make(chan struct{})
	<-ch

	return nil
}
