// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dto

import (
	"github.com/pkg/errors"
)

const (
	EDT_CUSTOM  = "CUSTOM"
	EDT_DEFAULT = "DEFAULT"
)

type EndpointDomainDto struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
}

type ServiceDetailDto struct {
	ServiceName     string              `json:"serviceName"`
	InnerAddress    string              `json:"innerAddress"`
	EndpointDomains []EndpointDomainDto `json:"endpointDomains"`
}

type RuntimeServiceReqDto struct {
	OrgId                 string             `json:"orgId"`
	ProjectId             string             `json:"projectID"`
	Env                   string             `json:"env"`
	ClusterName           string             `json:"clusterName"`
	RuntimeId             string             `json:"runtimeID"`
	RuntimeName           string             `json:"runtimeName"`
	ReleaseId             string             `json:"releaseId"`
	ServiceGroupNamespace string             `json:"serviceGroupNamespace"`
	ProjectNamespace      string             `json:"projectNamespace"`
	ServiceGroupName      string             `json:"serviceGroupName"`
	AppId                 string             `json:"appID"`
	AppName               string             `json:"appName"`
	Services              []ServiceDetailDto `json:"services"`
	UseApigw              *bool              `json:"useApigw"`
}

func (dto EndpointDomainDto) CheckValid() error {
	if dto.Domain == "" {
		return errors.New("empty domain")
	}
	if dto.Type != EDT_CUSTOM && dto.Type != EDT_DEFAULT {
		return errors.Errorf("invalid domain type:%s", dto.Type)
	}
	return nil
}

func (dto ServiceDetailDto) CheckValid() error {
	if dto.ServiceName == "" || dto.InnerAddress == "" {
		return errors.Errorf("invalid service req dto:%+v", dto)
	}
	for _, domain := range dto.EndpointDomains {
		err := domain.CheckValid()
		if err != nil {
			return err
		}
	}
	return nil
}

func (dto RuntimeServiceReqDto) CheckValid() error {
	if dto.ProjectId == "" || dto.OrgId == "" || dto.Env == "" || dto.ClusterName == "" ||
		dto.RuntimeId == "" || dto.RuntimeName == "" || dto.AppId == "" ||
		dto.UseApigw == nil {
		return errors.Errorf("invalid runtime req dto:%+v", dto)
	}
	for _, service := range dto.Services {
		err := service.CheckValid()
		if err != nil {
			return errors.WithMessagef(err, "invalid runtime req dto:%+v ", dto)
		}
	}
	return nil
}
