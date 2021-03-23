package snippetsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

type SnippetSvc struct {
	dbClient *dbclient.Client
	bdl      *bundle.Bundle
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle) *SnippetSvc {
	s := SnippetSvc{}
	s.dbClient = dbClient
	s.bdl = bdl
	return &s
}
