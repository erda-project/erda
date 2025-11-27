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
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"

	"github.com/pyroscope-io/pyroscope/pkg/util/bytesize"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
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
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct {
	MaxAudioSizeStr         string            `json:"maxAudioSize" yaml:"maxAudioSize"`
	MaxAudioSize            bytesize.ByteSize `json:"-" yaml:"-"`
	SupportedAudioFileTypes []string          `json:"supportedAudioFileTypes" yaml:"supportedAudioFileTypes"`
	DefaultOpenAIAudioModel string            `json:"defaultOpenAIAudioModel" yaml:"defaultOpenAIAudioModel"`
}

var Creator filter_define.RequestRewriterCreator = func(name string, configJSON json.RawMessage) filter_define.ProxyRequestRewriter {
	var ctx Context
	if err := yaml.Unmarshal(configJSON, &ctx); err != nil {
		panic(err)
	}
	maxAudioSize, err := bytesize.Parse(ctx.MaxAudioSizeStr)
	if err != nil {
		panic(fmt.Errorf("failed to parse maxAudioSize: %s", err))
	}
	ctx.MaxAudioSize = maxAudioSize
	return &ctx
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// parse multipart form
	if err := pr.In.ParseMultipartForm(int64(f.MaxAudioSize)); err != nil {
		return err
	}
	_, fileInfo, err := pr.In.FormFile(formBodyFieldFile)
	if err != nil {
		return err
	}

	// get prompt
	prompt := pr.In.FormValue(formBodyFieldPrompt)

	// check file type
	ext := strings.TrimPrefix(filepath.Ext(fileInfo.Filename), ".")
	supported := strutil.InSlice(ext, f.SupportedAudioFileTypes)
	if !supported {
		return fmt.Errorf("unsupported audio file type: %s (support: %v)", ext, f.SupportedAudioFileTypes)
	}

	// check file size
	if fileInfo.Size > int64(f.MaxAudioSize) {
		return fmt.Errorf("audio file size exceeds limit: %s (max: %s)", bytesize.ByteSize(fileInfo.Size), f.MaxAudioSize)
	}

	// set model name for audio request
	model := ctxhelper.MustGetModel(pr.In.Context())
	var modelName string = model.Name
	if customModelName := model.Metadata.Public["model_name"]; customModelName != nil {
		modelName = customModelName.GetStringValue()
	}
	pr.In.MultipartForm.Value[formBodyFieldModel] = []string{modelName}

	// reconstruct body
	newMultipartBody, err := reconstructMultipartBody(pr.In)
	if err != nil {
		return err
	}
	if err := body_util.SetBody(pr.Out, io.NopCloser(newMultipartBody)); err != nil {
		return fmt.Errorf("failed to set request body: %w", err)
	}

	audithelper.Note(pr.In.Context(), "prompt", prompt)
	audithelper.Note(pr.In.Context(), "audio.file_name", fileInfo.Filename)
	audithelper.Note(pr.In.Context(), "audio.file_size", bytesize.ByteSize(fileInfo.Size))
	audithelper.Note(pr.In.Context(), "audio.file_headers", fileInfo.Header)

	return nil
}

func reconstructMultipartBody(r *http.Request) (*bytes.Buffer, error) {
	originalBoundary, err := getMultipartBoundary(r.Header.Get(httperrorutil.HeaderKeyContentType))
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

func getMultipartBoundary(contentType string) (string, error) {
	if contentType == "" {
		return "", fmt.Errorf("empty Content-Type header")
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("failed to parse Content-Type %q: %w", contentType, err)
	}

	if !strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		return "", fmt.Errorf("unexpected media type %q, want multipart/*", mediaType)
	}

	boundary, ok := params["boundary"]
	if !ok || boundary == "" {
		return "", fmt.Errorf("no boundary parameter found in Content-Type %q", contentType)
	}

	// Be tolerant of optional quotes/spaces around the boundary value
	boundary = strings.Trim(boundary, "\" ")

	return boundary, nil
}
