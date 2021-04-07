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

package apistructs

import (
	"strings"
	"time"
)

const (
	EnvOpenapiTokenForActionBootstrap = "DICE_OPENAPI_TOKEN_FOR_ACTION_BOOTSTRAP"
	EnvOpenapiToken                   = "DICE_OPENAPI_TOKEN"
)

const (
	MetadataTypeDiceFile = "DiceFile"
)

type MetadataLevel string

var (
	MetadataLevelError MetadataLevel = "ERROR"
	MetadataLevelWarn  MetadataLevel = "WARN"
	MetadataLevelInfo  MetadataLevel = "INFO"
)

type (
	MetadataField struct {
		Name     string            `json:"name"`
		Value    string            `json:"value"`
		Type     string            `json:"type,omitempty"`
		Optional bool              `json:"optional,omitempty"`
		Labels   map[string]string `json:"labels,omitempty"`
		Level    MetadataLevel     `json:"level,omitempty"`
	}

	Metadata []MetadataField

	MetadataFieldType string
)

func (field MetadataField) GetLevel() MetadataLevel {
	if field.Level != "" {
		return field.Level
	}
	// judge by prefix
	idx := strings.Index(field.Name, ".")
	prefix := ""
	if idx != -1 {
		prefix = field.Name[:idx]
	} else {
		prefix = field.Name
	}
	switch MetadataLevel(strings.ToUpper(prefix)) {
	case MetadataLevelError:
		return MetadataLevelError
	case MetadataLevelWarn:
		return MetadataLevelWarn
	case MetadataLevelInfo:
		return MetadataLevelInfo
	}

	// fallback
	return MetadataLevelInfo
}

func (metadata Metadata) DedupByName() Metadata {
	tmp := make(map[string]struct{})
	dedup := make(Metadata, 0)
	for _, each := range metadata {
		if _, ok := tmp[each.Name]; ok {
			continue
		}
		tmp[each.Name] = struct{}{}
		dedup = append(dedup, each)
	}
	return dedup
}

// FilterNoErrorLevel filter by field level, return collection of NotErrorLevel and ErrorLevel.
func (metadata Metadata) FilterNoErrorLevel() (notErrorLevel, errorLevel Metadata) {
	for _, field := range metadata {
		if field.GetLevel() == MetadataLevelError {
			errorLevel = append(errorLevel, field)
			continue
		}
		notErrorLevel = append(notErrorLevel, field)
	}
	return
}

type ActionCallback struct {
	// show in stdout
	Metadata Metadata `json:"metadata"`

	Errors []ErrorResponse `json:"errors"`

	// machine stat
	MachineStat *PipelineTaskMachineStat `json:"machineStat,omitempty"`

	// behind
	PipelineID     uint64 `json:"pipelineID"`
	PipelineTaskID uint64 `json:"pipelineTaskID"`
}

const (
	ActionCallbackTypeLink             = "link"
	ActionCallbackRuntimeID            = "runtimeID"
	ActionCallbackOperatorID           = "operatorID"
	ActionCallbackReleaseID            = "releaseID"
	ActionCallbackPublisherID          = "publisherID"
	ActionCallbackPublishItemID        = "publishItemID"
	ActionCallbackPublishItemVersionID = "publishItemVersionID"
	ActionCallbackQaID                 = "qaID"
)

// detail
type ActionDetailResponse struct {
	Header
	Data interface{} `json:"data"`
}

// list
type ActionListResponse struct {
	Header
	Data interface{} `json:"data"`
}

type ActionCreateResponse struct {
	Header
	Data *ActionItem `json:"data"`
}

type ActionQueryResponse struct {
	Header
	Data []*ActionItem `json:"data"`
}

type ActionCreateRequest struct {
	// action名
	Name string `json:"name"`

	// action版本
	Version string `json:"version" `

	// spec yml 内容
	SpecSrc string `json:"specSrc" `

	// 源action镜像地址
	ImageSrc string `json:"imageSrc"`
}

type ActionSetStatusResponse struct {
	Header
}

type ActionItem struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	SpecSrc   string    `json:"specSrc"`
	Spec      string    `json:"spec"`
	ImageSrc  string    `json:"imageSrc"`
	Image     string    `json:"image"`
	IsDefault int       `json:"isDefault"`
	Desc      string    `json:"desc"`
	CreatedAt time.Time `json:"createdAt"`
}
