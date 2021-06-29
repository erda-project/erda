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

package task

import (
	"io"

	"github.com/erda-project/erda-infra/base/logs"
)

// WorkerManager .
type WorkerManager interface {
	Put(key int64, w Worker)
	Remove(key int64)
}

type simpleWorkerManager struct {
	worker map[int64]Worker
	log    logs.Logger
}

func newSimpleWorkerManager(log logs.Logger) WorkerManager {
	return &simpleWorkerManager{
		worker: make(map[int64]Worker),
		log:    log,
	}
}

func (p *simpleWorkerManager) Put(key int64, w Worker) {
	// remove worker if already exists
	p.Remove(key)
	p.worker[key] = w
}

func (p *simpleWorkerManager) Remove(key int64) {
	if w, ok := p.worker[key]; ok {
		if c, ok := w.(io.Closer); ok {
			if err := c.Close(); err != nil {
				p.log.Errorf("fail to close worker of key(%d)", key)
			}
		}
		delete(p.worker, key)
	}
}
