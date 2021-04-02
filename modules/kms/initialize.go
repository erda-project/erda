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
