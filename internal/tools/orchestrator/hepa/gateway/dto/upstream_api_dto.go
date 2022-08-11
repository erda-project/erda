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

package dto

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/hepa/legacy_upstream/pb"
)

type UpstreamApiDto struct {
	Path        string      `json:"path"`
	GatewayPath string      `json:"gatewayPath"`
	Method      string      `json:"method"`
	Address     string      `json:"address"`
	IsInner     bool        `json:"isInner"`
	IsCustom    bool        `json:"isCustom"`
	Doc         interface{} `json:"doc"`
	Name        string      `json:"name"`
	Domain      string      `json:"domain"`
}

func FromUpstreamApi(api *pb.UpstreamApi) UpstreamApiDto {
	return UpstreamApiDto{
		Path:        api.Path,
		GatewayPath: api.GatewayPath,
		Method:      api.Method,
		Address:     api.Address,
		IsInner:     api.IsInner,
		IsCustom:    api.IsCustom,
		Doc:         api.Doc.AsInterface(),
		Name:        api.Name,
		Domain:      api.GetDomain(),
	}
}

// Init adjusts *UpstreamApiDot.
// *UpstreamApiDto.Path is required to prefix "/";
// *UpstreamApiDto.Address is required to prefix "http";
// To set default *UpstreamApiDto.Name;
// To set default *UpstreamApiDto.GatewayPath.
func (dto *UpstreamApiDto) Init() error {
	if dto.Path == "" {
		return errors.Errorf("invalid path: %s", "path is empty")
	}
	if dto.Path[0] != '/' {
		return errors.Errorf("invalid path: %s", "missing prefix '/'")
	}
	if dto.Address == "" {
		return errors.Errorf("invalid address: %s", "address is empty")
	}
	if !strings.HasPrefix(dto.Address, "http") {
		return errors.Errorf("invalid address: %s", "scheme should be http or https")
	}
	if len(dto.Name) == 0 {
		dto.Name = dto.Method + dto.Path
	}
	if dto.GatewayPath == "" {
		dto.GatewayPath = dto.Path
	}

	return nil
}
