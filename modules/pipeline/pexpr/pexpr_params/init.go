package pexpr_params

import (
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

var dbClient *dbclient.Client

func Initialize(client *dbclient.Client) {
	dbClient = client
}
