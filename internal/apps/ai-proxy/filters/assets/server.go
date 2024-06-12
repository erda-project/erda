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

package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name         = "assets"
	AssetFileDir = "/tmp/ai-proxy/assets/" // auto clean by system when container restart
)

var (
	_ reverseproxy.RequestFilter = (*Filter)(nil)
)

var assetsMap = map[string]string{} // k: uuid, v: filepath
var lock = sync.Mutex{}

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
	if err := os.MkdirAll(AssetFileDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create assets dir: %v", err))
	}
}

type Filter struct {
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Filter{}, nil
}

func (f *Filter) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// get uuid from path
	pathMatcher := ctx.Value(reverseproxy.CtxKeyPathMatcher{}).(*route.PathMatcher)
	uuid := pathMatcher.Values["uuid"]
	if uuid == "" {
		return reverseproxy.Intercept, fmt.Errorf("missing uuid in path")
	}

	// find in memory
	lock.Lock()
	assetFilePath, ok := assetsMap[uuid]
	lock.Unlock()
	if !ok {
		return reverseproxy.Intercept, fmt.Errorf("asset not found")
	}
	// read file
	assetFile, err := os.OpenFile(assetFilePath, os.O_RDONLY, 0)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to open asset file: %v", err)
	}
	assetFileInfo, err := assetFile.Stat()
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to get asset file info: %v", err)
	}
	// write response for asset download
	w.Header().Set(httputil.HeaderKeyContentType, "application/octet-stream")
	w.Header().Set(httputil.HeaderKeyContentDisposition, fmt.Sprintf("attachment; filename=%s", getFileDisplayName(assetFilePath)))
	// must set, or we will get "Missing Content-Length of multimodal url" error
	w.Header().Set(httputil.HeaderKeyContentLength, strconv.FormatInt(assetFileInfo.Size(), 10))
	// copy file to response
	_, err = io.Copy(w, assetFile)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to copy asset file to response: %v", err)
	}
	return reverseproxy.Intercept, nil
}
