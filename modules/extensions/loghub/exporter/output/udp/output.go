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

package udp

import (
	"fmt"
	"net"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/extensions/loghub/exporter"
)

type config struct {
	Addr    string        `file:"addr"`
	Timeout time.Duration `file:"timeout"`
}

type provider struct {
	C   *config
	exp exporter.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.exp = ctx.Service("logs-exporter-base").(exporter.Interface)
	return nil
}

func (p *provider) Start() error {
	return p.exp.NewConsumer(p.newOutput)
}

func (p *provider) Close() error { return nil }

func (p *provider) newOutput(i int) (exporter.Output, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", p.C.Addr)
	if err != nil {
		return nil, fmt.Errorf("fail to resolve udp addr %s", err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	return &udpOutput{udpAddr, conn, p.C.Timeout}, nil
}

type udpOutput struct {
	addr    *net.UDPAddr
	conn    *net.UDPConn
	timeout time.Duration
}

func (o *udpOutput) Write(logkey string, data []byte) error {
	o.conn.SetWriteDeadline(time.Now().Add(o.timeout))
	_, err := o.conn.Write(data)
	if err != nil {
		conn, err := net.DialUDP("udp", nil, o.addr)
		if err != nil {
			return err
		}
		o.conn = conn
	}
	return err
}

func init() {
	servicehub.Register("logs-exporter-udp", &servicehub.Spec{
		Services:     []string{"logs-exporter-udp"},
		Dependencies: []string{"logs-exporter-base"},
		Description:  "logs export to udp",
		ConfigFunc: func() interface{} {
			return &config{Timeout: 30 * time.Second}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
	servicehub.Register("logs-exporter-logstash-udp", &servicehub.Spec{
		Services:     []string{"logs-exporter-logstash-udp"},
		Dependencies: []string{"logs-exporter-base"},
		Description:  "logs export to udp",
		ConfigFunc: func() interface{} {
			return &config{Timeout: 30 * time.Second}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
