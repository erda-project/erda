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

package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/buraksezer/consistent"
)

type Type string

func (t Type) String() string { return string(t) }

var (
	Official  Type = "official"
	Candidate Type = "candidate"
)

// Worker .
type Worker interface {
	consistent.Member
	json.Marshaler
	json.Unmarshaler

	ID() ID
	CreatedAt() time.Time
	Type() Type
	SetType(typ Type)
	Handle(ctx context.Context, data interface{})
}

type OpFunc func(*defaultWorker)

func WithID(id ID) OpFunc {
	return func(dw *defaultWorker) {
		dw.id = id
	}
}
func WithHandler(h handler) OpFunc {
	return func(dw *defaultWorker) {
		dw.handlers = append(dw.handlers, h)
	}
}

func New(ops ...OpFunc) Worker {
	var dw defaultWorker
	dw.id = NewID()
	dw.createdAt = time.Now()
	dw.typ = Candidate
	for _, op := range ops {
		op(&dw)
	}
	return &dw
}

func NewFromBytes(bytes []byte) (Worker, error) {
	var dw defaultWorker
	if err := dw.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}
	return &dw, nil
}

type defaultWorker struct {
	id        ID
	createdAt time.Time
	typ       Type
	handlers  []handler

	lock sync.Mutex
}

type handler func(ctx context.Context, data interface{})

func (dw *defaultWorker) UnmarshalJSON(bytes []byte) error { return json.Unmarshal(bytes, dw) }
func (dw *defaultWorker) MarshalJSON() ([]byte, error)     { return json.Marshal(dw) }
func (dw *defaultWorker) String() string                   { return string(dw.ID()) }
func (dw *defaultWorker) ID() ID                           { return dw.id }
func (dw *defaultWorker) CreatedAt() time.Time             { return dw.createdAt }
func (dw *defaultWorker) Type() Type                       { return dw.typ }
func (dw *defaultWorker) SetType(typ Type)                 { dw.typ = typ }
func (dw *defaultWorker) Handle(ctx context.Context, data interface{}) {
	dw.lock.Lock()
	var handlers []handler
	copy(handlers, dw.handlers)
	dw.lock.Unlock()

	finishChan := make(chan struct{}, len(handlers))
	finishedNum := 0
	wctx, wcancel := context.WithCancel(ctx)
	defer wcancel()
	for _, h := range dw.handlers {
		go func(h handler) {
			h(wctx, data)
			finishChan <- struct{}{}
		}(h)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-finishChan:
			finishedNum++
			if finishedNum == len(handlers) {
				close(finishChan)
				return
			}
		}
	}
}
