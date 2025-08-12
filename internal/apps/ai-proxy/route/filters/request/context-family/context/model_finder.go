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
	openai_v1_models "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/openai-v1-models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

// ModelIdentifier contains model identification information
type ModelIdentifier struct {
	ID   string // Model UUID (if provided)
	Name string // Model name
}

// ModelFinder defines the interface for finding model identifiers
type ModelFinder interface {
	Find(req *http.Request, ctx *ModelFinderContext) (*ModelIdentifier, error)
}

// ModelFinderContext provides additional context information
type ModelFinderContext struct {
	// Original request context, used to get path parameters, etc.
	RequestContext interface{}
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

func (f *HeaderFinder) Find(req *http.Request, ctx *ModelFinderContext) (*ModelIdentifier, error) {
	// Check Model ID first
	if modelID := req.Header.Get(vars.XAIProxyModelId); modelID != "" {
		return &ModelIdentifier{ID: modelID}, nil
	}

	// Check Model Name
	if modelName := req.Header.Get(vars.XAIProxyModel); modelName != "" {
		// Try to parse [ID:xxx] format
		if uuid := openai_v1_models.ParseModelUUIDFromDisplayName(modelName); uuid != "" {
			return &ModelIdentifier{ID: uuid, Name: modelName}, nil
		}
		return &ModelIdentifier{Name: modelName}, nil
	}

	if modelName := req.Header.Get(vars.XAIProxyModelName); modelName != "" {
		// Try to parse [ID:xxx] format
		if uuid := openai_v1_models.ParseModelUUIDFromDisplayName(modelName); uuid != "" {
			return &ModelIdentifier{ID: uuid, Name: modelName}, nil
		}
		return &ModelIdentifier{Name: modelName}, nil
	}

	return nil, nil
}

// PathFinder finds model identifier from path parameters
type PathFinder struct{}

func (f *PathFinder) Find(req *http.Request, ctx *ModelFinderContext) (*ModelIdentifier, error) {
	// Try to get path parameters from request context
	modelName, ok := ctxhelper.GetPathParam(req.Context(), "model")
	if !ok {
		return nil, nil
	}

	// Try to parse [ID:xxx] format
	if uuid := openai_v1_models.ParseModelUUIDFromDisplayName(modelName); uuid != "" {
		return &ModelIdentifier{ID: uuid, Name: modelName}, nil
	}

	return &ModelIdentifier{Name: modelName}, nil
}

// BodyFinder finds model identifier from request body
type BodyFinder struct{}

func (f *BodyFinder) Find(req *http.Request, ctx *ModelFinderContext) (*ModelIdentifier, error) {
	var finder BodyModelFinder
	var method, path string

	if req.URL != nil {
		method = req.Method
		path = req.URL.Path
	}

	// 1. First check if there is a special finder
	if customFinder := getCustomBodyModelNameFinderByMethodAndPath(method, path); customFinder != nil {
		modelName, err := customFinder.FindModelName(req, "model")
		if err != nil {
			return nil, err
		}
		if modelName != "" {
			// Try to parse [ID:xxx] format
			if uuid := openai_v1_models.ParseModelUUIDFromDisplayName(modelName); uuid != "" {
				return &ModelIdentifier{ID: uuid, Name: modelName}, nil
			}
			return &ModelIdentifier{Name: modelName}, nil
		}
		return nil, nil
	}

	// 2. Select standard finder based on Content-Type
	contentType := req.Header.Get("Content-Type")
	finder = getStandardFinderByContentType(contentType)

	// 3. Check if there is a special field name
	fieldKey := "model" // default
	if customField := getCustomBodyModelFieldByPathAndMethod(method, path); customField != "" {
		fieldKey = customField
	}

	modelName, err := finder.FindModelName(req, fieldKey)
	if err != nil {
		return nil, err
	}

	if modelName != "" {
		// Try to parse [ID:xxx] format
		if uuid := openai_v1_models.ParseModelUUIDFromDisplayName(modelName); uuid != "" {
			return &ModelIdentifier{ID: uuid, Name: modelName}, nil
		}
		return &ModelIdentifier{Name: modelName}, nil
	}

	return nil, nil
}

// findModelIdentifier unified model identifier lookup function
func findModelIdentifier(req *http.Request, requestCtx interface{}) (*ModelIdentifier, error) {
	// Create model finder context
	ctx := &ModelFinderContext{
		RequestContext: requestCtx,
	}

	// Search in priority order
	finders := []ModelFinder{
		&HeaderFinder{},
		&PathFinder{},
		&BodyFinder{},
	}

	for _, finder := range finders {
		identifier, err := finder.Find(req, ctx)
		if err != nil {
			return nil, err
		}
		if identifier != nil && (identifier.ID != "" || identifier.Name != "") {
			return identifier, nil
		}
	}

	return nil, nil
}
