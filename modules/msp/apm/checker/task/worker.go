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
	"context"
	"fmt"
	"io"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/plugins"
)

// Worker .
type Worker interface {
	Run(ctx context.Context)
}

var (
	defaultStrategy         = "periodic"
	defaultPeriodicInterval = 30 * time.Second
)

func NewWorker(log logs.Logger, c *pb.Checker, ctx plugins.Context, handler plugins.Handler) (Worker, error) {
	strategy := c.Config["strategy"]
	if len(strategy) == 0 {
		strategy = defaultStrategy
	}
	switch strategy {
	case "periodic":
		return newPeriodicWorker(log, c.Config, ctx, handler)
	case "cron":
		// TODO .
	}
	return nil, fmt.Errorf("strategy %q not support", strategy)
}

type periodicWorker struct {
	log      logs.Logger
	ctx      plugins.Context
	handler  plugins.Handler
	interval time.Duration
	closeCh  chan struct{}
}

func newPeriodicWorker(log logs.Logger, cfg map[string]string, ctx plugins.Context, handler plugins.Handler) (*periodicWorker, error) {
	interval := defaultPeriodicInterval
	if i := cfg["interval"]; len(i) > 0 {
		i, err := time.ParseDuration(i)
		if err != nil {
			return nil, fmt.Errorf("invalid interval %s", err)
		}
		if i <= 0 {
			return nil, fmt.Errorf("invalid interval %d", int64(i))
		}
		interval = i
	}
	return &periodicWorker{
		log:      log,
		ctx:      ctx,
		handler:  handler,
		interval: interval,
		closeCh:  make(chan struct{}),
	}, nil
}

func (w *periodicWorker) Run(ctx context.Context) {
	for {
		err := w.handler.Do(w.ctx)
		if err != nil {
			w.log.Error(err)
		}
		select {
		case <-time.After(w.interval):
		case <-w.closeCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (w *periodicWorker) Close() error {
	close(w.closeCh)
	if c, ok := w.handler.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// cronWorker TODO
type cronWorker struct {
}
