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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pyroscope-io/pyroscope/pkg/util/bytesize"
	"sigs.k8s.io/yaml"

	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "context-audio"

	formBodyFieldFile           = "file"
	formBodyFieldPrompt         = "prompt"
	formBodyFieldResponseFormat = "response_format"
	formBodyFieldModel          = "model"
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
	DefaultOpenAIAudioModel string            `json:"defaultOpenAIAudioModel" yaml:"defaultOpenAIAudioModel"`
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

	// force set model for openai
	modelProvider, _ := ctxhelper.GetModelProvider(ctx)
	if modelProvider.Type == modelproviderpb.ModelProviderType_OpenAI.String() {
		r.MultipartForm.Value[formBodyFieldModel] = []string{f.Config.DefaultOpenAIAudioModel}
	}

	// reconstruct body
	newMultipartBody, err := reconstructMultipartBody(r)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	infor.SetBody(io.NopCloser(newMultipartBody), int64(newMultipartBody.Len()))

	// put audio info into context
	audioInfo := ctxhelper.AudioInfo{
		FileName:    fileInfo.Filename,
		FileSize:    bytesize.ByteSize(fileInfo.Size),
		FileHeaders: fileInfo.Header,
	}
	ctxhelper.PutAudioInfo(ctx, audioInfo)

	return reverseproxy.Continue, nil
}

func reconstructMultipartBody(r *http.Request) (*bytes.Buffer, error) {
	originalBoundary, err := getMultipartBoundary(r.Header.Get(httputil.HeaderKeyContentType))
	if err != nil {
		return nil, err
	}

	newBody := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(newBody)
	if err := multipartWriter.SetBoundary(originalBoundary); err != nil {
		return nil, err
	}
	for field, files := range r.MultipartForm.File {
		for _, file := range files {
			part, err := multipartWriter.CreateFormFile(field, file.Filename)
			if err != nil {
				return nil, err
			}
			fileReader, err := file.Open()
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(part, fileReader); err != nil {
				return nil, err
			}
		}
	}
	for field, values := range r.MultipartForm.Value {
		for _, value := range values {
			part, err := multipartWriter.CreateFormField(field)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(part, strings.NewReader(value)); err != nil {
				return nil, err
			}
		}
	}
	if err := multipartWriter.Close(); err != nil {
		return nil, err
	}
	return newBody, nil
}

func getMultipartBoundary(contentType string) (boundary string, err error) {
	defer func() {
		if r := recover(); r != nil {
			boundary = ""
			err = fmt.Errorf("failed to get multipart boundary: %v", r)
		}
		if boundary == "" {
			err = fmt.Errorf("failed to get multipart boundary")
		}
	}()
	boundary = strings.SplitN(contentType, ";", 2)[1]
	boundary = strings.SplitN(boundary, "=", 2)[1]
	return boundary, nil
}
