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

package converter

import (
	"strings"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
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

type MetadataFieldType string

func MetadataFieldGetLevel(field *commonpb.MetadataField) MetadataLevel {
	if field == nil {
		return MetadataLevelInfo
	}
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

func MetadataDedupByName(metadata []*commonpb.MetadataField) []*commonpb.MetadataField {
	tmp := make(map[string]struct{})
	dedup := make([]*commonpb.MetadataField, 0)
	for _, each := range metadata {
		if _, ok := tmp[each.Name]; ok {
			continue
		}
		tmp[each.Name] = struct{}{}
		dedup = append(dedup, each)
	}
	return dedup
}

// MetadataFilterNoErrorLevel filter by field level, return collection of NotErrorLevel and ErrorLevel.
func MetadataFilterNoErrorLevel(metadata []*commonpb.MetadataField) (notErrorLevel, errorLevel []*commonpb.MetadataField) {
	for _, field := range metadata {
		if MetadataFieldGetLevel(field) == MetadataLevelError {
			errorLevel = append(errorLevel, field)
			continue
		}
		notErrorLevel = append(notErrorLevel, field)
	}
	return
}
