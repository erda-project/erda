// Package conf defines config options.
package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

// Conf define config from envs.
type Conf struct {
	ListenAddr    string             `env:"LISTEN_ADDR" default:":3082"`
	Debug         bool               `env:"DEBUG" default:"false"`
	KmsStoreKind  kmstypes.StoreKind `env:"KMS_STORE_KIND" default:"ETCD"`
	EtcdEndpoints string             `env:"ETCD_ENDPOINTS" required:"false"`
}

var cfg Conf

// Load load config from envs.
func Load() {
	envconf.MustLoad(&cfg)

	if cfg.KmsStoreKind == kmstypes.StoreKind_ETCD {
		if len(cfg.EtcdEndpoints) == 0 {
			panic("missing env ETCD_ENDPOINTS while KMS_STORE_KIND is ETCD")
		}
	}
}

// ListenAddr return ListenAddr option.
func ListenAddr() string {
	return cfg.ListenAddr
}

// Debug
func Debug() bool {
	return cfg.Debug
}

func KmsStoreKind() kmstypes.StoreKind {
	return cfg.KmsStoreKind
}

func EtcdEndpoints() string {
	return cfg.EtcdEndpoints
}
