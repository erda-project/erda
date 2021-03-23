package crondsvc

import (
	"sync"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/jsonstore"
)

type CrondSvc struct {
	crond *cron.Cron

	mu       *sync.Mutex
	dbClient *dbclient.Client
	bdl      *bundle.Bundle
	js       jsonstore.JsonStore
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle, js jsonstore.JsonStore) *CrondSvc {
	d := CrondSvc{}
	d.crond = cron.New()
	d.mu = &sync.Mutex{}
	d.dbClient = dbClient
	d.bdl = bdl
	d.js = js
	return &d
}
