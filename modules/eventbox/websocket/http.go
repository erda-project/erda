// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
