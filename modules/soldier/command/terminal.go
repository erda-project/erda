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

package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/soldier/settings"
)

const username = "erda"

var (
	uid, gid uint32
	homeDir  string
)

func init() {
	u, err := user.Lookup(username)
	if err != nil {
		logrus.Fatalln(err)
	}
	f := func(s string) (uint32, error) {
		u, err := strconv.ParseUint(u.Uid, 10, 32)
		return uint32(u), err
	}
	uid, err = f(u.Uid)
	if err != nil {
		logrus.Fatalln(err)
	}
	gid, err = f(u.Gid)
	if err != nil {
		logrus.Fatalln(err)
	}
	homeDir = u.HomeDir
}

func setUser(cmd *exec.Cmd) {
	cmd.Dir = homeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uid, Gid: gid}}
	for i, n := 0, len(cmd.Env); i < n; i++ {
		if strings.HasPrefix(cmd.Env[i], "HOME=") ||
			strings.HasPrefix(cmd.Env[i], "USER=") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			i--
			n--
		}
	}
	cmd.Env = append(cmd.Env, "HOME="+homeDir, "USER="+username)
}

func setEnv(cmd *exec.Cmd) {
	var term, lang bool
	for i, n := 0, len(cmd.Env); i < n; i++ {
		if strings.HasPrefix(cmd.Env[i], "TERM=") {
			term = true
		} else if strings.HasPrefix(cmd.Env[i], "LANG=") {
			lang = true
		}
	}
	if !term {
		cmd.Env = append(cmd.Env, "TERM=xterm-256color")
	}
	if !lang {
		cmd.Env = append(cmd.Env, "LANG=en_US.UTF-8")
	}
}

func isPort(i int) bool {
	return i >= 1 && i <= 65535
}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  BufferSize,
	WriteBufferSize: BufferSize,
	Subprotocols:    []string{"soldier"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Action struct {
	Env  []string         `json:"env"`
	Name string           `json:"name"`
	Args *json.RawMessage `json:"args"`
}

func (a Action) TerminalContext(ctx context.Context) (cmd *exec.Cmd, err error) {
	var b []string
	switch a.Name {
	case "bash":
		if a.Args != nil {
			err = errors.New("args unwanted")
		}
	case "docker":
		b, err = a.Docker()
	case "ssh":
		b, err = a.SSH()
	default:
		err = errors.New("name unsupported")
	}
	if err == nil {
		cmd = exec.CommandContext(ctx, a.Name, b...)
		cmd.Env = a.Env
		if settings.ForwardPort >= 0 {
			setUser(cmd)
		}
		setEnv(cmd)
	}
	return
}

type ActionDocker struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Container string `json:"container"`
}

func (a Action) Docker() ([]string, error) {
	if a.Args == nil {
		return nil, errors.New("args required")
	}
	v := ActionDocker{
		Port: 2375,
	}
	if err := json.Unmarshal(*a.Args, &v); err != nil {
		return nil, err
	}
	if v.Host == "" {
		return nil, errors.New("host required")
	}
	if !isPort(v.Port) {
		return nil, errors.New("port invalid")
	}
	if v.Container == "" {
		return nil, errors.New("container required")
	}
	args := make([]string, 0, 6)
	if settings.ForwardPort >= 0 {
		args = append(args, "-H", fmt.Sprintf("tcp://%s:%d", v.Host, v.Port))
	}
	args = append(args, "exec", "-it", v.Container,
		"/bin/sh", "-c", "if [ -f /bin/bash ]; then /bin/bash; else /bin/sh; fi")
	return args, nil
}

type ActionSSH struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
}

func (a Action) SSH() ([]string, error) {
	if a.Args == nil {
		return nil, errors.New("args required")
	}
	v := ActionSSH{
		Port: 22,
	}
	if err := json.Unmarshal(*a.Args, &v); err != nil {
		return nil, err
	}
	if v.Host == "" {
		return nil, errors.New("host required")
	}
	if !isPort(v.Port) {
		return nil, errors.New("port invalid")
	}
	if v.User == "" {
		return nil, errors.New("user required")
	}
	return []string{
		"-o", "StrictHostKeyChecking=no",
		"-p", strconv.Itoa(v.Port),
		v.User + "@" + v.Host,
	}, nil
}

var dialer = websocket.Dialer{ //
	ReadBufferSize:   BufferSize,
	WriteBufferSize:  BufferSize,
	HandshakeTimeout: 10 * time.Second,
	Subprotocols:     []string{"soldier"},
}

func Terminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Debugln(err)
		return
	}
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	defer func() {
		logrus.Infoln("terminal disconnected:", addr)
	}()
	logrus.Infoln("terminal connected:", addr)

	var a Action
	{
		t, b, err := conn.ReadMessage()
		if err != nil {
			logrus.Debugln(err)
			return
		}
		if t != websocket.TextMessage {
			return
		}
		err = json.Unmarshal(b, &a)
		if err != nil {
			logrus.Debugln(err)
			return
		}
		if settings.ForwardPort > 0 && a.Name == "docker" { //
			if a.Args == nil {
				logrus.Debugln("args required")
				return
			}
			var v ActionDocker
			err := json.Unmarshal(*a.Args, &v)
			if err != nil {
				logrus.Debugln(err)
				return
			}
			if v.Host == "" {
				logrus.Debugln("host required")
				return
			}
			u := fmt.Sprintf("ws://%s:%d/api/terminal", v.Host, settings.ForwardPort)
			c, _, err := dialer.Dial(u, nil)
			if err != nil {
				logrus.Errorln(err)
				return
			}
			defer c.Close()
			err = c.WriteMessage(t, b)
			if err != nil {
				logrus.Errorln(err)
				return
			}
			rc := make(chan struct{})
			go func() {
				defer close(rc)
				for {
					t, b, err := conn.ReadMessage()
					if err != nil {
						logrus.Debugln(err)
						return
					}
					err = c.WriteMessage(t, b)
					if err != nil {
						logrus.Debugln(err)
						return
					}
				}
			}()
			wc := make(chan struct{})
			go func() {
				defer close(wc)
				for {
					t, b, err := c.ReadMessage()
					if err != nil {
						logrus.Debugln(err)
						return
					}
					err = conn.WriteMessage(t, b)
					if err != nil {
						logrus.Debugln(err)
						return
					}
				}
			}()
			select {
			case <-rc:
			case <-wc:
			}
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd, err := a.TerminalContext(ctx)
	if err != nil {
		logrus.Debugln(err)
		return
	}
	slave, err := pty.Start(cmd)
	if err != nil {
		logrus.Debugln(err)
		return
	}
	defer slave.Close()

	var mu sync.Mutex
	write := func(b []byte) error {
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteMessage(websocket.BinaryMessage, b)
	}

	go func() {
		defer cancel()
		b := make([]byte, 1+BufferSize)
		b[0] = Output
		for {
			n, err := slave.Read(b[1:])
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

	go func() {
		defer cancel()
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
			case Input:
				_, err := slave.Write(b[1:])
				if err != nil {
					logrus.Debugln(err)
					return
				}
			case GetSize:
				if len(b) != 1 {
					return
				}
				v, err := pty.GetsizeFull(slave)
				if err != nil {
					logrus.Debugln(err)
					return
				}
				b, err := json.Marshal(v)
				if err != nil {
					logrus.Debugln(err)
					return
				}
				err = write(append([]byte{Size}, b...))
				if err != nil {
					logrus.Debugln(err)
					return
				}
			case SetSize:
				var v pty.Winsize
				err := json.Unmarshal(b[1:], &v)
				if err != nil {
					logrus.Debugln(err)
					return
				}
				err = pty.Setsize(slave, &v)
				if err != nil {
					logrus.Debugln(err)
					return
				}
			default:
				return
			}
		}
	}()

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		err := cmd.Wait()
		if err != nil {
			logrus.Debugln(err)
		}
	}()

	t := time.Tick(time.Minute)
	for {
		select {
		case <-ch:
			return
		case <-settings.ExitChan:
			return
		case <-t:
			if err := write([]byte{Ping}); err != nil {
				logrus.Debugln(err)
				return
			}
		}
	}
}
