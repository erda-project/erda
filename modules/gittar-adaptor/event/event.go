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

// Package event 事件信息
package event

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/gittar-adaptor/conf"
)

// Event 事件结构体
type Event struct {
	EventData
	bdl *bundle.Bundle
	// 事件集合，key为事件名称
	Events map[string]*apistructs.GittarRegisterHookRequest
}

// EventData 注册事件的数据结构
type EventData struct {
	// 事件名称
	Name string
	// 回调的路径
	Path string
	// 是否push触发，默认为true
	IsPush bool
}

// Option Event 类型定义
type Option func(*Event)

// New Event
func New(options ...Option) *Event {
	r := &Event{
		Events: make(map[string]*apistructs.GittarRegisterHookRequest),
	}

	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(f *Event) {
		f.bdl = bdl
	}
}

// Register 注册 gittar 事件
func (e *Event) Register(ed *EventData) error {
	if e.Events == nil {
		e.Events = make(map[string]*apistructs.GittarRegisterHookRequest)
	}

	if ed.Name == "" {
		return errors.New("need event name")
	}

	e.Events[ed.Name] = &apistructs.GittarRegisterHookRequest{
		Name:       ed.Name,
		URL:        fmt.Sprint("http://", conf.SelfAddr(), ed.Path),
		PushEvents: ed.IsPush,
	}

	err := e.bdl.RegisterGittarHook(*e.Events[ed.Name])
	return err
}
