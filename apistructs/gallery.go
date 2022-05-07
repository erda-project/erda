// Copyright (c) 2022 Terminus, Inc.
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
)

const (
	OpusTypeExtensionAction  OpusType = "erda/extension/action"
	OpusTypeExtensionAddon   OpusType = "erda/extension/addon"
	OpusTypeArtifactsProject OpusType = "erda/artifacts/project"

	OpusLevelSystem OpusLevel = "sys"
	OpusLevelOrg    OpusLevel = "sys"

	PutOnOpusModeAppend   PutOnOpusMode = "append"
	PutOnOpusModeOverride PutOnOpusMode = "override"

	LangUnkown      = "unknown"
	LangEn          = "en"
	LangEnUs   Lang = "en_us"
	LangZh          = "zh"
	LangZhCn   Lang = "zh_cn"
)

var (
	OpusTypes = map[OpusType]string{
		OpusTypeExtensionAction:  "Action",
		OpusTypeExtensionAddon:   "Addon",
		OpusTypeArtifactsProject: "Erda Artifacts",
	}

	LangTypes = map[Lang]string{
		LangUnkown: "unknown",
		LangEn:     "English",
		LangEnUs:   "English",
		LangZh:     "中文",
		LangZhCn:   "中文",
	}
)

type OpusType string

func (o OpusType) String() string {
	return string(o)
}

func (o OpusType) Equal(s string) bool {
	return strings.EqualFold(o.String(), s)
}

type OpusLevel string

func (o OpusLevel) String() string {
	return string(o)
}

func (o OpusLevel) Equal(s string) bool {
	return strings.EqualFold(o.String(), s)
}

type PutOnOpusMode string

func (o PutOnOpusMode) String() string {
	return string(o)
}

func (o PutOnOpusMode) Equal(s string) bool {
	return strings.EqualFold(o.String(), s)
}

type Lang string

func (o Lang) String() string {
	return string(o)
}

func (o Lang) Equal(s string) bool {
	return strings.EqualFold(strings.ReplaceAll(o.String(), "-", "_"), strings.ReplaceAll(s, "-", "_"))
}
