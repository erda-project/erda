package websocket

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/igm/sockjs-go.v2/sockjs"

	"github.com/erda-project/erda/modules/eventbox/input"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type WebsocketHTTP struct {
	m *websocket.Manager
	s chan struct{}
}

func New() (*WebsocketHTTP, error) {
	m, err := websocket.New()
	if err != nil {
		return nil, err
	}

	return &WebsocketHTTP{
		m: m,
		s: make(chan struct{}),
	}, nil
}

func (w *WebsocketHTTP) Start(handler input.Handler) error {
	w.m.Start()
	logrus.Info("Websocket start() done")
	// blocking
	<-w.s
	return nil
}

func (w *WebsocketHTTP) Stop() error {
	w.m.Stop()
	w.s <- struct{}{}
	logrus.Info("Websocket stop() done")
	return nil
}

func (w *WebsocketHTTP) Name() string {
	return "Websocket"
}

func (w *WebsocketHTTP) HTTPHandle(s sockjs.Session) {
	w.m.Handle(s)
}
