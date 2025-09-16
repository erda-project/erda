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

package temp_assets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name         = "temp-assets"
	AssetFileDir = "/tmp/ai-proxy/assets/" // auto clean by system when container restart
)

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

var assetsMap = map[string]string{} // k: uuid, v: filepath
var lock = sync.Mutex{}

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
	if err := os.MkdirAll(AssetFileDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create assets dir: %v", err))
	}
}

type Filter struct {
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Filter{}
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()
	// get uuid from path
	pathMatcher := ctxhelper.MustGetPathMatcher(ctx)
	uuid := pathMatcher.Values["uuid"]
	if uuid == "" {
		return http_error.NewHTTPError(http.StatusBadRequest, "missing uuid in path")
	}

	assetsMap["1234"] = "/Users/sfwn/Downloads/tg_image_83647306.png"

	// find in memory
	lock.Lock()
	assetFilePath, ok := assetsMap[uuid]
	lock.Unlock()
	if !ok {
		return fmt.Errorf("asset not found")
	}
	// read file
	assetFile, err := os.OpenFile(assetFilePath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open asset file: %v", err)
	}
	assetFileInfo, err := assetFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get asset file info: %v", err)
	}

	// construct response
	respHeader := http.Header{}
	respHeader.Set(httperrorutil.HeaderKeyContentType, "application/octet-stream")
	respHeader.Set(httperrorutil.HeaderKeyContentDisposition, fmt.Sprintf("attachment; filename=%s", getFileDisplayName(assetFilePath)))
	// must set, or we will get "Missing Content-Length of multimodal url" error
	respHeader.Set(httperrorutil.HeaderKeyContentLength, strconv.FormatInt(assetFileInfo.Size(), 10))

	resp := &http.Response{
		StatusCode:    http.StatusOK,
		Header:        respHeader,
		Body:          assetFile, // copy file to response body
		ContentLength: assetFileInfo.Size(),
		Request:       pr.Out,
	}

	// trigger filter-generated response
	transports.TriggerRequestFilterGeneratedResponse(pr.Out, resp)

	return nil
}
