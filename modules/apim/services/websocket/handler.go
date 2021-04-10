package websocket

import (
	"io"

	"github.com/erda-project/erda/apistructs"
)

type ResponseWriter interface {
	io.Writer
}

type Handler func(w ResponseWriter, r *apistructs.WebsocketRequest) error
