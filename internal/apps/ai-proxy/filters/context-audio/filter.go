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
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/pyroscope-io/pyroscope/pkg/util/bytesize"
	"io"
	"net/http"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name                        = "context-audio"
	formBodyFieldFile           = "file"
	formBodyFieldPrompt         = "prompt"
	formBodyFieldResponseFormat = "response_format"

	allowedAudioResponseFormat = "json"
)

var (
	_ reverseproxy.RequestFilter = (*Context)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Context struct {
	Config *Config
}

type Config struct {
	MaxAudioSizeStr         string            `json:"maxAudioSize" yaml:"maxAudioSize"`
	MaxAudioSize            bytesize.ByteSize `json:"-" yaml:"-"`
	SupportedAudioFileTypes []string          `json:"supportedAudioFileTypes" yaml:"supportedAudioFileTypes"`
}

func New(configJSON json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}
	maxAudioSize, err := bytesize.Parse(cfg.MaxAudioSizeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maxAudioSize: %s", err)
	}
	cfg.MaxAudioSize = maxAudioSize
	return &Context{Config: &cfg}, nil
}

func (f *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	r := infor.Request()

	// recover body finally
	originalBody := infor.BodyBuffer(true)
	defer func() {
		infor.SetBody(io.NopCloser(originalBody), int64(originalBody.Len()))
	}()

	// parse multipart form
	if err := r.ParseMultipartForm(int64(f.Config.MaxAudioSize)); err != nil {
		return reverseproxy.Intercept, err
	}
	_, fileInfo, err := r.FormFile(formBodyFieldFile)
	if err != nil {
		return reverseproxy.Intercept, err
	}

	// get prompt
	prompt := r.FormValue(formBodyFieldPrompt)
	ctxhelper.PutUserPrompt(ctx, prompt)

	// check file type
	ext := strings.TrimPrefix(filepath.Ext(fileInfo.Filename), ".")
	supported := strutil.InSlice(ext, f.Config.SupportedAudioFileTypes)
	if !supported {
		return reverseproxy.Intercept, fmt.Errorf("unsupported audio file type: %s (support: %v)", ext, f.Config.SupportedAudioFileTypes)
	}

	// check file size
	if fileInfo.Size > int64(f.Config.MaxAudioSize) {
		return reverseproxy.Intercept, fmt.Errorf("audio file size exceeds limit: %s (max: %s)", bytesize.ByteSize(fileInfo.Size), f.Config.MaxAudioSize)
	}

	// put audio info into context
	audioInfo := ctxhelper.AudioInfo{
		FileName:    fileInfo.Filename,
		FileSize:    bytesize.ByteSize(fileInfo.Size),
		FileHeaders: fileInfo.Header,
	}
	ctxhelper.PutAudioInfo(ctx, audioInfo)

	return reverseproxy.Continue, nil
}
