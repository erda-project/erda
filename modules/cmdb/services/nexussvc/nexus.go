package nexussvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/pkg/encryption"
)

// NexusSvc nexus 操作封装
type NexusSvc struct {
	db  *dao.DBClient
	bdl *bundle.Bundle

	rsaCrypt *encryption.RsaCrypt
}

type Option func(*NexusSvc)

func New(options ...Option) *NexusSvc {
	svc := &NexusSvc{}
	for _, op := range options {
		op(svc)
	}

	return svc
}

func WithDBClient(db *dao.DBClient) Option {
	return func(svc *NexusSvc) {
		svc.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *NexusSvc) {
		svc.bdl = bdl
	}
}

func WithRsaCrypt(rsaCrypt *encryption.RsaCrypt) Option {
	return func(svc *NexusSvc) {
		svc.rsaCrypt = rsaCrypt
	}
}
