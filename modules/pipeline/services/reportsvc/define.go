package reportsvc

import (
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

type ReportSvc struct {
	dbClient *dbclient.Client
}

func New(ops ...Option) *ReportSvc {
	var svc ReportSvc

	for _, op := range ops {
		op(&svc)
	}

	return &svc
}

type Option func(*ReportSvc)

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(svc *ReportSvc) {
		svc.dbClient = dbClient
	}
}
