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
	"sync"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

type ImageInfo struct {
	ImageQuality string `json:"imageQuality"`
	ImageSize    string `json:"imageSize"`
	ImageStyle   string `json:"imageStyle"`
}

func GetImageInfo(ctx context.Context) (*ImageInfo, bool) {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyImageInfo{})
	if !ok || value == nil {
		return nil, false
	}
	info, ok := value.(*ImageInfo)
	if !ok {
		return nil, false
	}
	return info, true
}

func PutImageInfo(ctx context.Context, info ImageInfo) {
	m := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyImageInfo{}, &info)
}
