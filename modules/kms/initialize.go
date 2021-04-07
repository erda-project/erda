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

// Package kms provides unified Key Management Service.
package kms

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/kms/conf"
	"github.com/erda-project/erda/modules/kms/endpoints"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/kms"
	"github.com/erda-project/erda/pkg/kms/stores/etcd"
)

// Initialize initialize and bootstrap the module.
func Initialize() error {
	conf.Load()

	// config logrus
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	kmsMgr, err := kms.GetManager(kms.WithStoreConfigs(map[string]string{
		etcd.EnvKeyEtcdEndpoints: conf.EtcdEndpoints(),
	}))
	if err != nil {
		return err
	}

	ep := endpoints.New(endpoints.WithKmsManager(kmsMgr))

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())

	if err := do(); err != nil {
		return err
	}

	return server.ListenAndServe()
}

func do() error {
	return nil
}
