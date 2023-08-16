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

package main

import (
	"embed"
	"io"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/swagger"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

//go:embed apispec.yaml
var apiSpec embed.FS

func getAPISpec(path, method string) *openapi3.Operation {
	f, err := apiSpec.Open("apispec.yaml")
	if err != nil {
		log.Fatalf("failed to open apispec.oas3: %v", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read apispec.oas3: %v", err)
	}
	sw, err := swagger.LoadFromData(b)
	if err != nil {
		log.Fatalf("failed to load apispec.oas3: %v", err)
	}
	if p, ok := sw.Paths[path]; ok {
		switch method {
		case http.MethodGet:
			return p.Get
		case http.MethodPost:
			return p.Post
		case http.MethodPut:
			return p.Put
		case http.MethodDelete:
			return p.Delete
		default:
			log.Fatalf("unsupported method: %s", method)
		}
	}
	return nil
}

func mustGetOneAPI() *apistructs.APIOperation {
	apiSummaries, err := bdl.SearchAPIOperations(orgID, userID, "Add a new pet to the store")
	if err != nil {
		log.Fatalf("failed to search api operations, err: %v", err)
	}
	if len(apiSummaries) == 0 {
		log.Fatal("no api operations found")
	}
	apiSummary := apiSummaries[0]
	apiDetail, err := bdl.GetAPIOperation(orgID, userID, apiSummary.ID)
	if err != nil {
		log.Fatalf("failed to get api operation, err: %v", err)
	}
	// fulfill by api spec
	apiOperation := getAPISpec(apiDetail.Path, apiDetail.Method)
	if apiOperation == nil {
		return apiDetail
	}
	//handleAPIOperations(apiDetail, apiOperation)
	return apiDetail
}

func handleAPIOperations(apiDetail *apistructs.APIOperation, operation *openapi3.Operation) {
	// 处理 parameters
	for _, p := range operation.Parameters {
		if p == nil || p.Value == nil || p.Value.Schema == nil || p.Value.Schema.Value == nil {
			continue
		}
		v := p.Value
		schema := v.Schema.Value
		parameter := apistructs.Parameter{
			Name:            v.Name,
			Description:     v.Description,
			AllowEmptyValue: v.AllowEmptyValue,
			AllowReserved:   v.AllowReserved,
			Deprecated:      v.Deprecated,
			Required:        v.Required,
			Type:            schema.Type,
			Enum:            schema.Enum,
			Default:         schema.Default,
			Example:         schema.Example,
		}
		switch v.In {
		case "query":
			apiDetail.Parameters = append(apiDetail.Parameters, &parameter)
		case "header":
			apiDetail.Headers = append(apiDetail.Headers, &parameter)
		}
	}

	// 处理 request body
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		v := operation.RequestBody.Value
		apiDetail.RequestBodyDescription = v.Description
		apiDetail.RequestBodyRequired = v.Required
		for mediaType, c := range v.Content {
			if c.Schema == nil || c.Schema.Value == nil {
				continue
			}
			if c.Schema.Value != nil {
				oas3.GenExampleFromExpandedSchema(httputil.ContentType(mediaType), c.Schema.Value)
			}
			apiDetail.RequestBody = append(apiDetail.RequestBody, &apistructs.RequestBody{
				MediaType: mediaType,
				Body:      c.Schema.Value,
			})
		}
	}

	// 处理 responses
	for statusCode, res := range operation.Responses {
		if res.Value == nil {
			continue
		}
		for mediaType, c := range res.Value.Content {
			if c.Schema == nil || c.Schema.Value == nil {
				continue
			}
			resp := apistructs.Response{
				StatusCode:  statusCode,
				MediaType:   mediaType,
				Description: "",
				Body:        c.Schema.Value,
			}
			if res.Value.Description != nil {
				resp.Description = *res.Value.Description
			}
		}
	}
}
