package models

import (
	"github.com/erda-project/erda/bundle"
)

type Service struct {
	db     *DBClient
	bundle *bundle.Bundle
}

func NewService(dbClient *DBClient, bundle *bundle.Bundle) *Service {
	s := Service{}
	s.db = dbClient
	s.bundle = bundle
	return &s
}
