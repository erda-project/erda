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

// Package ops Core components of multi-cloud management platform
package ops

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
)

const serviceOps = "ops"

type provider struct{}

// Service Declare what services the provider provides
func (p *provider) Service() []string { return []string{serviceOps} }

// Dependencies Return which services the provider depends on
func (p *provider) Dependencies() []string { return []string{} }

// Description Describe information about this provider
func (p *provider) Description() string {
	return "Core components of multi-cloud management platform."
}

// Creator Return a provider creator
func (p *provider) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

// TODO: refactor
// Init Initialize the provider to run
func (p *provider) Init(ctx servicehub.Context) error { return nil }

// Start Start the provider
func (p *provider) Start() error {
	logrus.Info("starting the ops provider...")
	return nil
}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	logrus.Info("ops provider is running...")
	return initialize()
}

// Close Close the provider
func (p *provider) Close() error {
	logrus.Info("ops provider is closing...")
	return nil
}

// TODO: refactor

func init() {
	servicehub.RegisterProvider(serviceOps, &provider{})
}
