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

package adapter

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/template"
)

type adapterService struct {
	p *provider
}

const (
	MSP_ENV_ID   = "msp_env_id"
	EnvCollector = "COLLECTOR_ADDR"
)

type InstrumentationLibrary struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Template string `json:"template"`
}

func (s *adapterService) GetInstrumentationLibrary(ctx context.Context, request *pb.GetInstrumentationLibraryRequest) (*pb.GetInstrumentationLibraryResponse, error) {
	result := &pb.GetInstrumentationLibraryResponse{
		Data: make([]*pb.InstrumentationLibrary, 0),
	}
	for k, v := range s.p.libraryMap {
		libraryList := &pb.InstrumentationLibrary{
			Strategy: k,
		}
		languages := make([]*pb.Language, 0)
		languageList, ok := v.([]interface{})
		if !ok {
			return nil, errors.NewInternalServerError(fmt.Errorf("instrumentation library language is invalidate,language is %v", v))
		}
		for _, v := range languageList {
			language := &pb.Language{
				Language: v.(string),
			}
			languages = append(languages, language)
		}
		libraryList.Languages = languages
		result.Data = append(result.Data, libraryList)
	}
	return result, nil
}

func (s *adapterService) GetInstrumentationLibraryDocs(ctx context.Context, request *pb.GetInstrumentationLibraryDocsRequest) (*pb.GetInstrumentationLibraryDocsResponse, error) {
	renderMap := map[string]string{
		MSP_ENV_ID:   request.ScopeId,
		EnvCollector: os.Getenv(EnvCollector),
	}
	file, err := os.Open(s.p.configFile)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer file.Close()
	decode := yaml.NewDecoder(file)
	libraryArr := make([]InstrumentationLibrary, 0)
	err = decode.Decode(&libraryArr)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	var data string
	for _, v := range libraryArr {
		if v.Language == request.Language && v.Name == request.Strategy {
			data = template.Render(v.Template, renderMap)
		}
	}
	return &pb.GetInstrumentationLibraryDocsResponse{
		Data: data,
	}, nil
}
