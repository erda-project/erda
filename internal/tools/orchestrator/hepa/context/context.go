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

package context

import (
	"context"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/sirupsen/logrus"
)

const (
	baseSkip = 1
	maxSkip  = 20
)

type LogContext struct {
	context.Context

	l        logger
	baseFunc string
}

func (c LogContext) Entry() (entry *logrus.Entry) {
	defer func() {
		formatter := new(logrus.TextFormatter)
		formatter.SortingFunc = sorting
		entry.Logger.SetFormatter(formatter)
	}()

	pc, _, _, ok := runtime.Caller(baseSkip)
	if !ok {
		switch t := c.l.(type) {
		case *logrus.Logger:
			return logrus.NewEntry(t)
		case *logrus.Entry:
			return t
		}
		return logrus.NewEntry(logrus.StandardLogger())
	}
	name := filepath.Base(runtime.FuncForPC(pc).Name())
	if c.baseFunc == "" || c.baseFunc == name {
		return c.l.WithField("stack", name)
	}
	for i := baseSkip + 1; i < baseSkip+maxSkip; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		curName := filepath.Base(runtime.FuncForPC(pc).Name())
		name = curName + " -> " + name
		if curName == c.baseFunc {
			break
		}
	}
	return c.l.WithField("stack", name)
}

func WithLoggerIfWithout(ctx context.Context, logger *logrus.Logger) *LogContext {
	return with(ctx, logger)
}

func WithEntryIfWithout(ctx context.Context, entry *logrus.Entry) *LogContext {
	return with(ctx, entry)
}

type logger interface {
	WithField(string, interface{}) *logrus.Entry
}

func with(ctx context.Context, logger logger) *LogContext {
	if t, ok := ctx.(*LogContext); ok {
		return t
	}

	if pc, _, _, ok := runtime.Caller(baseSkip + 1); ok {
		baseName := filepath.Base(runtime.FuncForPC(pc).Name())
		c := &LogContext{
			Context:  ctx,
			l:        logger,
			baseFunc: baseName,
		}
		return c
	}
	return &LogContext{
		Context: ctx,
		l:       logger,
	}
}

func sorting(args []string) {
	asc := []string{"time", "level", "error", "msg"}
	desc := []string{"line", "stack"}
	sort.Slice(args, func(i, j int) bool {
		for _, s := range asc {
			if args[i] == s {
				return true
			}
			if args[j] == s {
				return false
			}
		}
		for _, s := range desc {
			if args[i] == s {
				return false
			}
			if args[j] == s {
				return true
			}
		}
		return args[i] < args[j]
	})
}
