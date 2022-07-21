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
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/core/hepa/legacy_upstream/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/pkg/strutil"
)

type UpstreamRegisterDto struct {
	Az            string           `json:"az"`
	UpstreamName  string           `json:"-"`
	DiceAppId     string           `json:"diceAppId"`
	DiceService   string           `json:"diceService"`
	RuntimeName   string           `json:"runtimeName"`
	RuntimeId     string           `json:"runtimeId"`
	AppName       string           `json:"appName"`
	ServiceAlias  string           `json:"serviceName"`
	OrgId         string           `json:"orgId"`
	ProjectId     string           `json:"projectId"`
	Env           string           `json:"env"`
	ApiList       []UpstreamApiDto `json:"apiList"`
	OldRegisterId *int             `json:"registerId"`
	RegisterId    string           `json:"registerTag"`
	PathPrefix    *string          `json:"pathPrefix"`
	Scene         string           `json:"scene"`
}

// FromUpstream 转换结构
func FromUpstream(u *pb.Upstream) *UpstreamRegisterDto {
	dto := &UpstreamRegisterDto{
		Az:           u.Az,
		UpstreamName: "",
		DiceAppId:    u.DiceAppId,
		DiceService:  u.DiceService,
		RuntimeName:  u.RuntimeName,
		RuntimeId:    u.RuntimeId,
		AppName:      u.AppName,
		ServiceAlias: u.ServiceName,
		OrgId:        u.OrgId,
		ProjectId:    u.ProjectId,
		Env:          strings.ToUpper(u.Env),
		ApiList:      nil,
		RegisterId:   u.RegisterTag,
		PathPrefix:   nil,
		Scene:        u.GetScene(),
	}
	if dto.Scene == "" {
		dto.Scene = orm.UnityScene
	}
	var apiList []UpstreamApiDto
	for _, api := range u.ApiList {
		apiList = append(apiList, FromUpstreamApi(api))
	}
	dto.ApiList = apiList
	if u.RegisterId != 0 {
		id := int(u.RegisterId)
		dto.OldRegisterId = &id
	}
	if u.PathPrefix != "" {
		dto.PathPrefix = &u.PathPrefix
	}
	return dto
}

// Init .
// concat *UpstreamRegisterDto.UpstreamName;
// concat *UpstreamRegisterDto.PathPrefix;
// generate *UpstreamRegisterDto.OldRegisterId;
// trim UpstreamApiDto from *UpstreamRegisterDto.ApiList if !UpstreamApiDto.IsCustom;
// set default UpstreamRegisterDto.ApiList if no custom UpstreamApiDto.
func (dto *UpstreamRegisterDto) Init() error {
	if err := dto.preCheck(); err != nil {
		return err
	}
	dto.UpstreamName = dto.AppName
	if dto.DiceService == "" {
		dto.DiceService = dto.ServiceAlias
	}
	if dto.ServiceAlias != "" {
		dto.UpstreamName += "/" + dto.ServiceAlias
	}
	if dto.RuntimeId != "" {
		dto.UpstreamName += "/" + dto.RuntimeId
	}
	if dto.RuntimeName == "" {
		if dto.PathPrefix == nil {
			path := strings.ToLower("/" + dto.UpstreamName)
			dto.PathPrefix = &path
		} else {
			path := "/" + strings.TrimPrefix(*dto.PathPrefix, "/")
			path = strings.TrimSuffix(path, "/")
			dto.PathPrefix = &path
		}
	} else {
		dto.PathPrefix = nil
	}
	if dto.OldRegisterId != nil && dto.RegisterId == "" {
		dto.RegisterId = strconv.Itoa(*dto.OldRegisterId)
	}
	if err := dto.AdjustAPIsDomains(); err != nil {
		return err
	}
	if dto.Scene == orm.HubScene && len(dto.ApiList) > 0 && dto.ApiList[0].Domain != "" {
		dto.UpstreamName = orm.HubSceneUpstreamNamePrefix + "/" + dto.UpstreamName
	}
	address := dto.ApiList[0].Address
	var apisM = make(map[string]struct{})
	for i := len(dto.ApiList) - 1; i >= 0; i-- {
		if _, ok := apisM[dto.ApiList[i].Path]; ok || !dto.ApiList[i].IsCustom {
			dto.ApiList = append(dto.ApiList[:i], dto.ApiList[i+1:]...)
		}
		apisM[dto.ApiList[i].Path] = struct{}{}
	}
	if len(dto.ApiList) == 0 {
		dto.ApiList = []UpstreamApiDto{
			{
				Path:        "/",
				GatewayPath: "/",
				Method:      "",
				Address:     address,
				IsInner:     false,
				Name:        "/",
			},
		}
	}
	return nil
}

func (dto *UpstreamRegisterDto) preCheck() error {
	switch dto.Scene {
	case "":
		dto.Scene = orm.UnityScene
	case orm.UnityScene, orm.HubScene:
	default:
		return errors.Errorf("you can only register routes to %s scene package", orm.UnityScene+"|"+orm.HubScene)
	}
	if dto.AppName == "" {
		return errors.Errorf("invalid appName: %s", "appName is empty")
	}
	if dto.OrgId == "" {
		return errors.Errorf("invalid orgID: %s", "orgID is empty")
	}
	if dto.ProjectId == "" {
		return errors.Errorf("invalid projectID: %s", "projectID is empty")
	}
	if len(dto.ApiList) == 0 {
		return errors.Errorf("invalid apiList: %s", "apiList is empty")
	}
	switch dto.Env {
	case "PROD", "STAGING", "TEST", "DEV":
	default:
		return errors.Errorf("invalid env, expects PROD,STAGING,TEST,DEV, got %s", dto.Env)
	}
	if dto.OldRegisterId == nil && dto.RegisterId == "" {
		return errors.New("invalid registerID")
	}
	return nil
}

func (dto *UpstreamRegisterDto) AdjustAPIsDomains() error {
	if dto.Scene == "" || dto.Scene == orm.UnityScene {
		for i := range dto.ApiList {
			dto.ApiList[i].Domain = ""
		}
		return nil
	}

	var domain string
	for i := range dto.ApiList {
		if dto.ApiList[i].Domain == "" {
			continue
		}
		domains := strings.Split(dto.ApiList[i].Domain, ",")
		for i, domain := range domains {
			domains[i] = strings.TrimSpace(domain)
		}
		domains = strutil.DedupSlice(domains)
		dto.ApiList[i].Domain = hepautils.SortJoinDomains(domains)
		if domain != "" && dto.ApiList[i].Domain != domain {
			return errors.Errorf("different domains in api list: ApiList[%v] %s, ApiList[%v] %s", i-1, domain, i, dto.ApiList[i].Domain)
		}
		domain = dto.ApiList[i].Domain
	}
	return nil
}
