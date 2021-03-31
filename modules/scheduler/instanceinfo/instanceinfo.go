package instanceinfo

import (
	"github.com/erda-project/erda/pkg/dbengine"
)

type Client struct {
	db *dbengine.DBEngine
}

func New(db *dbengine.DBEngine) *Client {
	// db.DB = db.DB.Debug()
	return &Client{
		db: db,
	}
}
