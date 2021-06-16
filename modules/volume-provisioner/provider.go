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

// config The definition of volume-provisioner config
type config struct {
	// ProvisionerNamespace Consistent with the namespace of the volume-provisioner daemonSet,
	// Usually injected directly from metadata.namespace
	ProvisionerNamespace string `env:"PROVISIONER_NAMESPACE" default:"default"`
	// LocalProvisionerName Name of local volume provisioner
	LocalProvisionerName string `env:"LOCAL_PROVISIONER_NAME" default:"dice/local-volume"`
	// NetProvisionerName Name of netData volume provisioner
	NetProvisionerName string `env:"NET_PROVISIONER_NAME" default:"dice/netdata-volume"`
	// LocalMatchLabel Match label for execute on specified pods
	LocalMatchLabel string `env:"LOCAL_MATCH_LABEL" default:"app=volume-provisioner"`
	// ModeEdge Used for edge computing,
	ModeEdge bool `env:"EDGE_MODE" default:"false"`
	// NodeName Used for edge computing, directory creation action on the specified edge nodeSite
	NodeName string `env:"NODE_NAME" default:""`
}

type provider struct {
	Cfg *config
}

// TODO: refactor
// Init Initialize the provider to run
func (p *provider) Init(ctx servicehub.Context) error { return nil }

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	logrus.Info("volumeProvisioner provider is running...")
	return initialize(p.Cfg, ctx)
}

// Close Close the provider
func (p *provider) Close() error {
	logrus.Info("volumeProvisioner provider is closing...")
	return nil
}

func init() {
	servicehub.Register("volume-provisioner", &servicehub.Spec{
		Services:    []string{"volume-provisioner"},
		Description: "This is a Persistent Volume Claim (PVC) provisioner for Kubernetes.",
		ConfigFunc:  func() interface{} { return &config{} },
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
