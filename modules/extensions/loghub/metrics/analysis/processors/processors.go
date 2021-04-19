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

package processors

import (
	"fmt"

	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

// Processor .
type Processor interface {
	Process(content string) (string, map[string]interface{}, error)
	Keys() []*metrics.FieldDefine
}

type processor struct {
	Processor
	tags map[string]string
}

var processors = make(map[string]func(metric string, cfg []byte) (Processor, error))

// RegisterProcessor .
func RegisterProcessor(typ string, creator func(metric string, cfg []byte) (Processor, error)) {
	processors[typ] = creator
}

// NewProcessor .
func NewProcessor(name, typ string, cfg []byte) (Processor, error) {
	creator, ok := processors[typ]
	if !ok {
		return nil, fmt.Errorf("processor %s not exist", typ)
	}
	return creator(name, cfg)
}

// Processors .
type Processors struct {
	ps map[string][]*processor
}

// New .
func New() *Processors {
	return &Processors{
		ps: make(map[string][]*processor),
	}
}

// Add .
func (ps *Processors) Add(key string, tags map[string]string, name, typ string, config []byte) error {
	creator, ok := processors[typ]
	if !ok {
		return fmt.Errorf("processor %s not exist", typ)
	}
	p, err := creator(name, config)
	if err != nil {
		return err
	}
	ps.ps[key] = append(ps.ps[key], &processor{
		tags:      tags,
		Processor: p,
	})
	return nil
}

// Find .
func (ps *Processors) Find(name, key string, tags map[string]string) []Processor {
	procs := ps.ps[key]
	var list []Processor
loop:
	for _, p := range procs {
		if len(p.tags) > 0 {
			for k, v := range p.tags {
				if tags[k] != v {
					continue loop
				}
			}
		}
		list = append(list, p)
	}
	return list
}
