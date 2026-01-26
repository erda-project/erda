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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// ModelFinder defines the interface for finding model identifiers
type ModelFinder interface {
	Find(req *http.Request) (string, error)
}

// BodyModelFinder defines the interface for finding model identifiers from request body
type BodyModelFinder interface {
	FindModelName(req *http.Request, fieldKey string) (string, error)
}

// JSONBodyFinder finds model name from JSON request body
type JSONBodyFinder struct{}

func (f *JSONBodyFinder) FindModelName(req *http.Request, fieldKey string) (string, error) {
	if !strings.HasPrefix(req.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		return "", nil
	}

	snapshotBody, err := body_util.SmartCloneBody(&req.Body, body_util.MaxSample)
	if err != nil {
		return "", fmt.Errorf("failed to clone request body: %v", err)
	}

	var reqBody map[string]interface{}
	if err := json.NewDecoder(snapshotBody).Decode(&reqBody); err != nil {
		return "", fmt.Errorf("failed to decode request body, err: %v", err)
	}

	if modelValue, exists := reqBody[fieldKey]; exists {
		if modelStr, ok := modelValue.(string); ok {
			return modelStr, nil
		}
	}

	return "", nil
}

// FormBodyFinder finds model name from form request body
type FormBodyFinder struct{}

func (f *FormBodyFinder) FindModelName(req *http.Request, fieldKey string) (string, error) {
	contentType := req.Header.Get(httperrorutil.HeaderKeyContentType)
	if !strings.Contains(contentType, string(httperrorutil.URLEncodedFormMime)) {
		return "", nil
	}

	snapshotReq, err := body_util.SafeCloneRequest(req, body_util.MaxSample)
	if err != nil {
		return "", fmt.Errorf("failed to clone request: %v", err)
	}

	// Parse form data
	if err := snapshotReq.ParseForm(); err != nil {
		return "", fmt.Errorf("failed to parse form: %v", err)
	}

	modelValue := snapshotReq.FormValue(fieldKey)
	return modelValue, nil
}

// MultipartFormBodyFinder finds model name from multipart form request body
type MultipartFormBodyFinder struct{}

func (f *MultipartFormBodyFinder) FindModelName(req *http.Request, fieldKey string) (string, error) {
	contentType := req.Header.Get(httperrorutil.HeaderKeyContentType)
	if !strings.Contains(contentType, "multipart/form-data") {
		return "", nil
	}

	// Parse multipart form if not already parsed
	if req.MultipartForm == nil {
		// Use a reasonable max memory limit for parsing (32MB)
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			return "", fmt.Errorf("failed to parse multipart form: %v", err)
		}
	}

	if req.MultipartForm != nil && req.MultipartForm.Value != nil {
		if modelValues, exists := req.MultipartForm.Value[fieldKey]; exists && len(modelValues) > 0 {
			return modelValues[0], nil
		}
	}

	return "", nil
}

// getCustomBodyModelNameFinderByMethodAndPath returns special model name finder based on request method and path
// Currently returns nil, indicating no special cases, use standard Content-Type judgment logic
func getCustomBodyModelNameFinderByMethodAndPath(method, path string) BodyModelFinder {
	// Currently no special cases require fully custom finder
	return nil
}

// getCustomBodyModelFieldByPathAndMethod returns special model field name based on request method and path
// Currently returns empty string, indicating use of default "model" field
func getCustomBodyModelFieldByPathAndMethod(method, path string) string {
	// Currently no special cases require APIs using non-"model" fields
	return ""
}

// getStandardFinderByContentType selects standard finder based on Content-Type
func getStandardFinderByContentType(contentType string) BodyModelFinder {
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return &FormBodyFinder{}
	} else if strings.Contains(contentType, "multipart/form-data") {
		return &MultipartFormBodyFinder{}
	} else {
		return &JSONBodyFinder{} // default
	}
}

// HeaderFinder finds model identifier from request headers
type HeaderFinder struct{}

func (f *HeaderFinder) Find(req *http.Request) (string, error) {
	return strutil.FirstNotEmpty(
		req.Header.Get(vars.XAIProxyModelId),
		req.Header.Get(vars.XAIProxyModel),
		req.Header.Get(vars.XAIProxyModelName),
	), nil
}

// PathFinder finds model identifier from path parameters
type PathFinder struct{}

func (f *PathFinder) Find(req *http.Request) (string, error) {
	// Try to get path parameters from request context
	modelName, ok := ctxhelper.GetPathParam(req.Context(), "model")
	if !ok {
		return "", nil
	}

	return modelName, nil
}

// QueryParamFinder finds model identifier from URL query parameters
type QueryParamFinder struct{}

func (f *QueryParamFinder) Find(req *http.Request) (string, error) {
	if req.URL == nil {
		return "", nil
	}
	return req.URL.Query().Get("model"), nil
}

// BodyFinder finds model identifier from request body
type BodyFinder struct{}

func (f *BodyFinder) Find(req *http.Request) (string, error) {
	var finder BodyModelFinder
	var method, path string

	if req.URL != nil {
		method = req.Method
		path = req.URL.Path
	}

	// 1. first check if there is a special finder
	if customFinder := getCustomBodyModelNameFinderByMethodAndPath(method, path); customFinder != nil {
		modelName, err := customFinder.FindModelName(req, "model")
		if err != nil {
			return "", err
		}
		if modelName != "" {
			return modelName, nil
		}
		return "", nil
	}

	// 2. select standard finder based on Content-Type
	contentType := req.Header.Get("Content-Type")
	finder = getStandardFinderByContentType(contentType)

	// 3. Check if there is a special field name
	fieldKey := "model" // default
	if customField := getCustomBodyModelFieldByPathAndMethod(method, path); customField != "" {
		fieldKey = customField
	}

	modelName, err := finder.FindModelName(req, fieldKey)
	if err != nil {
		return "", err
	}

	if modelName != "" {
		return modelName, nil
	}

	return "", nil
}

// findModelIdentifier unified model identifier lookup function
func findModelIdentifier(req *http.Request) (string, error) {
	// search in priority order
	finders := []ModelFinder{
		&HeaderFinder{},
		&PathFinder{},
		&QueryParamFinder{},
		&BodyFinder{},
	}

	for _, finder := range finders {
		identifier, err := finder.Find(req)
		if err != nil {
			return "", err
		}
		if identifier != "" {
			return identifier, nil
		}
	}

	return "", nil
}
