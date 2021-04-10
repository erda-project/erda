package apidocsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/apim/dbclient"
)

type Service struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

type Option func(service *Service)

func New(options ...Option) *Service {
	var service Service
	for _, op := range options {
		op(&service)
	}
	return &service
}

func WithDBClient(db *dbclient.DBClient) Option {
	return func(service *Service) {
		service.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(service *Service) {
		service.bdl = bdl
	}
}
