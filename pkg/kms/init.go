package kms

import (
	// plugins
	_ "github.com/erda-project/erda/pkg/kms/plugins/dicekms"

	// stores
	_ "github.com/erda-project/erda/pkg/kms/stores/etcd"
)
