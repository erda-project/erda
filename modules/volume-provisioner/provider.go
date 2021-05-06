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

	"github.com/erda-project/erda-infra/base/servicehub"
)

const serviceVolumeProvisioner = "volume-provisioner"

// define Represents the definition of provider and provides some information
type define struct{}

// Service Declare what services the provider provides
func (d *define) Service() []string { return []string{serviceVolumeProvisioner} }

// Dependencies Return which services the provider depends on
func (d *define) Dependencies() []string { return []string{} }

// Description Describe information about this provider
func (d *define) Description() string {
	return "This is a Persistent Volume Claim (PVC) provisioner for Kubernetes."
}

// Creator Return a provider creator
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

// Config Return an instance representing the configuration
func (d *define) Config() interface{} { return &config{} }

type provider struct {
	Cfg *config
}

// config The definition of volume-provisioner config
type config struct {
	// ProvisionerNamespace Consistent with the namespace of the volume-provisioner daemonSet,
	// Usually injected directly from metadata.namespace
	ProvisionerNamespace string `env:"PROVISIONER_NAMESPACE" default:"default"`
	// LocalProvisionerName Name of local volume provisioner
	LocalProvisionerName string `env:"LOCAL_PROVISIONER_NAME" default:"erda/local-volume"`
	// NetProvisionerName Name of netData volume provisioner
	NetProvisionerName string `env:"NET_PROVISIONER_NAME" default:"erda/netdata-volume"`
	// LocalMatchLabel Match label for execute on specified pods
	LocalMatchLabel string `env:"LOCAL_MATCH_LABEL" default:"app=volume-provisioner"`
	// ModeEdge Used for edge computing,
	ModeEdge bool `env:"EDGE_MODE" default:"false"`
	// NodeName Used for edge computing, directory creation action on the specified edge nodeSite
	NodeName string `env:"NODE_NAME" default:""`
}

// TODO: refactor
// Init Initialize the provider to run
func (p *provider) Init(ctx servicehub.Context) error { return nil }

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	logrus.Info("volumeProvisioner provider is running...")
	return initialize(p.Cfg)
}

// Close Close the provider
func (p *provider) Close() error {
	logrus.Info("volumeProvisioner provider is closing...")
	return nil
}

func init() {
	servicehub.RegisterProvider(serviceVolumeProvisioner, &define{})
}
