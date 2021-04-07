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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataField_GetLevel(t *testing.T) {
	field := MetadataField{Name: "Error.web"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: "ERROR.web"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: " ERROR.web"}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())

	field = MetadataField{Name: "ERROR"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: "WARN.."}
	assert.Equal(t, MetadataLevelWarn, field.GetLevel())

	field = MetadataField{Name: "warn"}
	assert.Equal(t, MetadataLevelWarn, field.GetLevel())

	field = MetadataField{Name: "INFO.x"}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())

	field = MetadataField{Name: "Error.x", Level: MetadataLevelInfo}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())
}
