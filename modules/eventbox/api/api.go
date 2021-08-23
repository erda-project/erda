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

package api

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/erda-project/erda/pkg/jsonstore"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type Notifier interface {
	Send(content interface{}, options ...OpOperation) error
	SendRaw(message *types.Message) error
}
type OpOperation func(*Op)

type Op struct {
	labels map[string]interface{}
	dest   map[string]interface{}
	sender string
}

type NotifierImpl struct {
	sender string
	labels map[string]interface{} // optional, can also assign `dest' as `Send' option
	dir    string
	js     jsonstore.JsonStore
}

func WithDest(dest map[string]interface{}) OpOperation {
	return func(op *Op) {
		op.dest = dest
	}
}

func WithLabels(labels map[string]interface{}) OpOperation {
	return func(op *Op) {
		op.labels = labels
	}
}

func WithSender(sender string) OpOperation {
	return func(op *Op) {
		op.sender = sender
	}
}

func genMessagePath(dir string, timestamp int64) string {
	return filepath.Join(dir, strconv.FormatInt(timestamp, 10))
}

func genMessage(sender string, content interface{}, timestamp int64, labels map[string]interface{}) (*types.Message, error) {
	return &types.Message{
		Sender:  sender,
		Content: content,
		Labels:  convert(labels),
		Time:    timestamp,
	}, nil
}

var js jsonstore.JsonStore

// dest 可以是nil， 可以在 Send 的 options 中指定 ( WithDest )
func New(sender string, dest map[string]interface{}) (Notifier, error) {
	var err error
	if js == nil {
		if js, err = jsonstore.New(); err != nil {
			return nil, err
		}
	}
	return &NotifierImpl{
		sender: sender,
		labels: dest,
		dir:    constant.MessageDir,
		js:     js,
	}, nil
}

// 将 message 写入 etcd
func (n *NotifierImpl) Send(content interface{}, options ...OpOperation) error {
	option := &Op{}
	for _, op := range options {
		op(option)
	}
	timestamp := time.Now().UnixNano()
	labels := n.labels
	sender := n.sender
	if option.dest != nil {
		labels = mergeMap(option.dest, labels)
	}
	if option.labels != nil {
		labels = mergeMap(option.labels, labels)
	}
	if option.sender != "" {
		sender = option.sender
	}

	message, err := genMessage(sender, content, timestamp, labels)
	if err != nil {
		return err
	}
	return n.SendRaw(message)
}

func (n *NotifierImpl) SendRaw(message *types.Message) error {
	messagePath := genMessagePath(n.dir, message.Time)
	ctx := context.Background()
	return n.js.Put(ctx, messagePath, message)
}

func convert(m map[string]interface{}) map[types.LabelKey]interface{} {
	m_ := map[types.LabelKey]interface{}{}
	for k, v := range m {
		m_[types.LabelKey(k)] = v
	}
	return m_
}

func mergeMap(main map[string]interface{}, minor map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range minor {
		m[k] = v
	}

	for k, v := range main {
		m[k] = v
	}
	return m
}
