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

package pipelineymlv1

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
)

func TestRegexp(t *testing.T) {
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	fmt.Println(re.ReplaceAllString("A string - with, awful characters$@!.", ""))
}

func TestGenerateTaskUUID(t *testing.T) {
	fmt.Println(GenerateTaskUUID(1, "- Makefile-1 ", 0, "-do !@#$ what i want to do-", "uuid"))
}

func TestApplyKVsWithPriority(t *testing.T) {
	// priority: e1 > e2 > e3
	e1 := map[string]string{
		"NAME": "linjun",
		"E1":   "V1",
	}
	e2 := map[string]string{
		"NAME": "name2",
		"E2":   "V2",
	}
	e3 := map[string]string{
		"NAME": "name3",
		"E3":   "V3",
	}

	result := ApplyKVsWithPriority(e3, e2, e1)
	require.Equal(t, 4, len(result))
	require.Equal(t, "linjun", result["NAME"])
}

func TestMap2MetadataFields(t *testing.T) {
	m := map[string]string{
		"oss.access.key": "1",
		"OSS_ACCESS_KEY": "2",
	}
	metas := Map2MetadataFields(m)
	require.True(t, len(metas) == 2)
}

func TestMetadataFields2Map(t *testing.T) {
	metas := []apistructs.MetadataField{
		{Name: "oss.access.key", Value: "1"},
		{Name: "OSS_ACCESS_KEY", Value: "2"},
	}
	m := MetadataFields2Map(metas)
	require.True(t, len(m) == 2)
	require.Equal(t, m["oss.access.key"], "1")
	require.Equal(t, m["OSS_ACCESS_KEY"], "2")
}
