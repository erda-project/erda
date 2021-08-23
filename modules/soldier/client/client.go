// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/containerd/console"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"
)

const BufferSize = 1024

const (
	Error   = '1'
	Ping    = '2'
	Pong    = '3'
	Input   = '4'
	Output  = '5'
	GetSize = '6'
	Size    = '7'
	SetSize = '8'
)

type Action struct {
	Env  []string    `json:"env"`
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

type ActionDocker struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Container string `json:"container"`
}

type ActionSSH struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

var dialer = websocket.Dialer{
	ReadBufferSize:  BufferSize,
	WriteBufferSize: BufferSize,
	Subprotocols:    []string{"soldier"},
}

func (a Action) Run(addr string) {
	u, err := url.Parse(addr)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return
	}
	u.Path = path.Join(u.Path, "/api/terminal")

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	defer conn.Close()

	master := console.Current()
	err = master.SetRaw()
	if err != nil {
		logrus.Errorln(err)
		return
	}
	defer master.Reset()

	b, err := json.Marshal(a)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	var mu sync.Mutex
	write := func(b []byte) error {
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteMessage(websocket.BinaryMessage, b)
	}

	rc := make(chan struct{})
	{
		ch := make(chan os.Signal, 1)
		defer close(ch)
		signal.Notify(ch, syscall.SIGWINCH)
		defer signal.Reset(syscall.SIGWINCH)
		go func() {
			defer close(rc)
			for {
				if _, ok := <-ch; !ok {
					return
				}
				size, err := master.Size()
				if err != nil {
					logrus.Debugln(err)
					return
				}
				logrus.Debugf("master size %dx%d", size.Width, size.Height)
				b, err := json.Marshal(pty.Winsize{
					Rows: size.Height,
					Cols: size.Width,
				})
				if err != nil {
					logrus.Debugln(err)
					return
				}
				err = write(append([]byte{SetSize}, b...))
				if err != nil {
					logrus.Debugln(err)
					return
				}
				err = write([]byte{GetSize})
				if err != nil {
					logrus.Debugln(err)
					return
				}
			}
		}()
		ch <- syscall.SIGWINCH
	}

	mc := make(chan struct{})
	go func() {
		defer close(mc)
		b := make([]byte, 1+BufferSize)
		b[0] = Input
		for {
			n, err := master.Read(b[1:])
			if err != nil {
				if err != io.EOF {
					logrus.Debugln(err)
				}
				if n > 0 {
					err = write(b[:n+1])
					if err != nil {
						logrus.Debugln(err)
					}
				}
				return
			}
			if n > 0 {
				err = write(b[:n+1])
				if err != nil {
					logrus.Debugln(err)
					return
				}
			}
		}
	}()

	sc := make(chan struct{})
	go func() {
		defer close(sc)
		for {
			t, b, err := conn.ReadMessage()
			if err != nil {
				logrus.Debugln(err)
				return
			}
			if t != websocket.BinaryMessage {
				return
			}
			if len(b) == 0 {
				return
			}
			switch b[0] {
			case Ping:
				if len(b) != 1 {
					return
				}
				err := write([]byte{Pong})
				if err != nil {
					logrus.Debugln(err)
					return
				}
			case Pong:
				if len(b) != 1 {
					return
				}
			case Error, Output:
				_, err := master.Write(b[1:])
				if err != nil {
					logrus.Debugln(err)
					return
				}
			case Size:
				var size pty.Winsize
				err := json.Unmarshal(b[1:], &size)
				if err != nil {
					logrus.Debugln(err)
					return
				}
				logrus.Debugf("slave size %dx%d", size.Cols, size.Rows)
			default:
				return
			}
		}
	}()

	select {
	case <-rc:
		return
	case <-mc:
		return
	case <-sc:
		return
	}
}
