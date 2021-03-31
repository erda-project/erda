package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/customhttp"
)

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

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

var passRegexp = regexp.MustCompile(`(?i)(?:pass|secret)[^\s\0]*=([^\s\0]+)`)

func hidePassEnv(b []byte) []byte {
	return passRegexp.ReplaceAllFunc(b, func(a []byte) []byte {
		if i := bytes.IndexByte(a, '='); i != -1 {
			for i += 2; i < len(a); i++ { // keep first
				a[i] = '*'
			}
		}
		return a
	})
}

func (k *Kubernetes) Terminal(namespace, podname, containername string, upperConn *websocket.Conn) {
	f := func(cols, rows uint16) (*websocket.Conn, error) {
		path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/exec", namespace, podname)
		s := `stty cols %d rows %d; s=/bin/sh; if [ -f /bin/bash ]; then s=/bin/bash; fi; `
		if conf.TerminalSecurity() {
			s += "if [ `id -un` != dice ]; then su -l dice -s $s; exit $?; fi; "
		}
		s += "$s"
		cmd := url.QueryEscape(fmt.Sprintf(s, cols, rows))
		query := "command=sh&command=-c&command=" + cmd + "&container=" + containername + "&stdin=1&stdout=1&tty=true"
		req, err := customhttp.NewRequest("GET", k.addr, nil)
		if err != nil {
			logrus.Errorf("failed to customhttp.NewRequest: %v", err)
			return nil, err
		}

		execURL := url.URL{
			Scheme:   "ws",
			Host:     req.URL.Host,
			Path:     path,
			RawQuery: query,
		}
		req.Header.Add("X-Portal-Websocket", "on")
		if req.Header.Get("X-Portal-Host") != "" {
			req.Header.Add("Host", req.Header.Get("X-Portal-Host"))
		}
		conn, _, err := websocket.DefaultDialer.Dial(execURL.String(), req.Header)
		if err != nil {
			logrus.Errorf("failed to connect to %s: %v", execURL.String(), err)
			return nil, err
		}
		return conn, nil
	}
	var conn *websocket.Conn
	var waitConn sync.WaitGroup
	waitConn.Add(1)
	var setsizeOnce sync.Once
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer func() {
			wait.Done()
			if conn != nil {
				conn.Close()
			}
			upperConn.Close()
		}()
		waitConn.Wait()
		for {
			if conn == nil {
				return
			}
			tp, m, err := conn.ReadMessage()
			if err != nil {
				return
			}
			m[0] = Output
			if conf.TerminalSecurity() {
				m = hidePassEnv(m)
			}
			if err := upperConn.WriteMessage(tp, m); err != nil {
				return
			}
		}
	}()
	go func() {
		defer func() {
			wait.Done()
			if conn != nil {
				conn.Close()
			}
			upperConn.Close()
		}()
		for {
			tp, m, err := upperConn.ReadMessage()
			if err != nil {
				return
			}
			switch m[0] {
			case Input:
				if conn == nil {
					continue
				}
				m[0] = 0 // k8s 协议, stdin = 0
				if err := conn.WriteMessage(tp, m); err != nil {
					return
				}
			case SetSize:
				var err error
				setsizeOnce.Do(func() {
					var v Winsize
					err = json.Unmarshal(m[1:], &v)
					if err != nil {
						return
					}
					conn, err = f(v.Cols, v.Rows)
					waitConn.Done()
					if err != nil {
						logrus.Errorf("failed to connect k8s exec ws: %v", err)
						return
					}
				})
				if err != nil {
					return
				}
			default:
				continue
			}
		}
	}()
	wait.Wait()
}
