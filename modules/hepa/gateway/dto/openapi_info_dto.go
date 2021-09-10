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

import "github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"

type Origin string

const (
	FROM_DICE    Origin = "dice"
	FROM_CUSTOM  Origin = "custom"
	FROM_DICEYML Origin = "diceyml"
	FROM_SHADOW  Origin = "shadow"
)

type OpenapiInfoDto struct {
	ApiId       string `json:"apiId"`
	CreateAt    string `json:"createAt"`
	DiceApp     string `json:"diceApp"`
	DiceService string `json:"diceService"`
	Origin      Origin `json:"origin"`
	Mutable     bool   `json:"mutable"`
	OpenapiDto
}

func (dto OpenapiInfoDto) ToEndpointApi() *pb.EndpointApi {
	ep := dto.OpenapiDto.ToEndpointApi()
	ep.ApiId = dto.ApiId
	ep.CreateAt = dto.CreateAt
	ep.DiceApp = dto.DiceApp
	ep.DiceService = dto.DiceService
	ep.Origin = string(dto.Origin)
	ep.Mutable = dto.Mutable
	return ep
}
