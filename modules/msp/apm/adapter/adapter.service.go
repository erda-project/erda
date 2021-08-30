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

	config2 "github.com/recallsong/go-utils/config"

	"github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/template"
)

type adapterService struct {
	p *provider
}

func (s *adapterService) GetInstrumentationLibrary(ctx context.Context, request *pb.GetInstrumentationLibraryRequest) (*pb.GetInstrumentationLibraryResponse, error) {
	result := &pb.GetInstrumentationLibraryResponse{
		Data: make([]*pb.InstrumentationLibrary, 0),
	}
	for k, v := range s.p.libraryMap {
		libraryList := &pb.InstrumentationLibrary{
			DisplayName: k,
			Strategy:    k,
		}
		languages := make([]*pb.Language, 0)
		languageList, ok := v.([]interface{})
		if !ok {
			return nil, errors.NewInternalServerError(fmt.Errorf("instrumentation library language is invalidate,language is %v", v))
		}
		for _, v := range languageList {
			language := &pb.Language{
				Language:    v.(string),
				DisplayName: v.(string),
			}
			languages = append(languages, language)
		}
		libraryList.Languages = languages
		result.Data = append(result.Data, libraryList)
	}
	return result, nil
}

func (s *adapterService) GetInstrumentationLibraryDocs(ctx context.Context, request *pb.GetInstrumentationLibraryDocsRequest) (*pb.GetInstrumentationLibraryDocsResponse, error) {
	data, err := config2.LoadFile(s.p.configFile)
	renderMap := map[string]string{
		"language": request.Language,
		"strategy": request.Strategy,
	}
	configFile := template.Render(string(data), renderMap)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetInstrumentationLibraryDocsResponse{
		Data: configFile,
	}, nil
}
