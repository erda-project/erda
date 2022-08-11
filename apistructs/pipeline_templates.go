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

package apistructs

import (
	"errors"
	"time"
)

type TemplateVersion int

const (
	TemplateVersionV1 = TemplateVersion(1)
	TemplateVersionV2 = TemplateVersion(2)
)

type SnippetQueryDetailsRequest struct {
	SnippetConfigs []SnippetDetailQuery `json:"snippetConfigs,omitempty"`
}
type SnippetQueryDetailsResponse struct {
	Header
	Data map[string]SnippetQueryDetail `json:"data,omitempty"`
}
type SnippetQueryDetail struct {
	Params  []*PipelineParam `json:"params,omitempty"`
	Outputs []string         `json:"outputs,omitempty"`
}

type SnippetDetailQuery struct {
	SnippetConfig
	Alias string `json:"alias,omitempty"` // 别名
}

type PipelineTemplateVersionGetRequest struct {
	ScopeType string `json:"scopeType"`
	ScopeID   string `json:"scopeID"`
	Name      string `json:"name"`
	Version   string `json:"version"`
}

type PipelineTemplateRenderRequest struct {
	ScopeType       string                 `json:"scopeType"`
	ScopeID         string                 `json:"scopeID"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Params          map[string]interface{} `json:"params"`
	Alias           string                 `json:"alias"`
	TemplateVersion TemplateVersion        `json:"renderVersion"`
}

type PipelineTemplateRenderSpecRequest struct {
	Spec            *PipelineTemplateSpec  `json:"spec"`
	Alias           string                 `json:"alias"`
	TemplateVersion TemplateVersion        `json:"renderVersion"`
	Params          map[string]interface{} `json:"params"`
}

type PipelineTemplateRenderResponse struct {
	Header
	Data PipelineTemplateRender `json:"data"`
}

type PipelineTemplateVersionGetResponse struct {
	Header
	Data PipelineTemplateVersion `json:"data"`
}

type PipelineTemplateRender struct {
	PipelineYaml string                   `json:"pipelineYaml"`
	Version      *PipelineTemplateVersion `json:"pipelineTemplateVersion"`
	Outputs      []SnippetFormatOutputs   `json:"outputs"`
}

type SnippetFormatOutputs struct {
	PreOutputName   string `json:"PreOutputName"`
	AfterOutputName string `json:"AfterOutputName"`
}

type PipelineTemplateVersion struct {
	ID         uint64    `json:"id"`
	TemplateId uint64    `json:"templateId"`
	Name       string    `json:"name"`
	Version    string    `json:"version"`
	Spec       string    `json:"spec"`
	Readme     string    `json:"readme"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type PipelineTemplateSpec struct {
	Name           string            `json:"name" yaml:"name"`
	Version        string            `json:"version" yaml:"version"`
	Desc           string            `json:"desc" yaml:"desc"`
	LogoUrl        string            `json:"logoUrl" yaml:"logo_url"`
	Params         []*PipelineParam  `json:"params" yaml:"params"`
	Outputs        []*PipelineOutput `json:"outputs" yaml:"outputs"`
	Template       string            `json:"template" yaml:"template"`
	DefaultVersion string            `json:"defaultVersion" yaml:"default_version"`
}

func (p *PipelineTemplateSpec) Check() error {
	if p.Name == "" {
		return errors.New("spec name can not empty")
	}

	if p.Template == "" {
		return errors.New("spec template can not empty")
	}

	//if p.Version == "" {
	//	return errors.New("spec version can not empty")
	//}

	if p.Params != nil {
		for _, v := range p.Params {
			if err := v.Check(); err != nil {
				return err
			}
		}
	}

	if p.Outputs != nil {
		for _, v := range p.Outputs {
			if err := v.Check(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (params *PipelineParam) Check() error {

	if params.Name == "" {
		return errors.New("params name can not empty")
	}

	return nil
}

func (output *PipelineOutput) Check() error {
	if output.Name == "" {
		return errors.New("outputs name can not empty")
	}

	if output.Ref == "" {
		return errors.New("outputs ref can not empty")
	}

	return nil
}
