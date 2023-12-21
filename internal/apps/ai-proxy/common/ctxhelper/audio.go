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

package ctxhelper

import (
	"context"
	"net/textproto"
	"sync"

	"github.com/pyroscope-io/pyroscope/pkg/util/bytesize"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

type AudioInfo struct {
	FileName    string               `json:"fileName"`
	FileSize    bytesize.ByteSize    `json:"fileSize"`
	FileHeaders textproto.MIMEHeader `json:"fileHeaders"`
}

func GetAudioInfo(ctx context.Context) (*AudioInfo, bool) {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyAudioInfo{})
	if !ok || value == nil {
		return nil, false
	}
	info, ok := value.(*AudioInfo)
	if !ok {
		return nil, false
	}
	return info, true
}

func PutAudioInfo(ctx context.Context, info AudioInfo) {
	m := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyAudioInfo{}, &info)
}
