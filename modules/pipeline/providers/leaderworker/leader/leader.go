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

package leader

import (
	"context"
)

type Leader interface {
	ID() worker.ID
	SetHandler(h handler)
}

type defaultLeader struct {
	id      worker.ID
	handler func(ctx context.Context)
}

func (dl *defaultLeader) ID() worker.ID {
	return dl.id
}

func (dl *defaultLeader) SetHandler(h handler) {
	dl.handler = h
}

type handler func(ctx context.Context)

func New(id worker.ID, h handler) Leader {
	var dl defaultLeader
	dl.id = id
	dl.handler = h
	return &dl
}
