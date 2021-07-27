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

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/base/converter"
)

func Test_EnablePipelineVolume(t *testing.T) {
	var tables = []struct {
		StorageConfig *basepb.StorageConfig
	}{
		{
			StorageConfig: &basepb.StorageConfig{
				EnableNFS:   false,
				EnableLocal: true,
			},
		},
		{
			StorageConfig: &basepb.StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
		},
		{
			StorageConfig: &basepb.StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
		},
	}
	for _, data := range tables {
		if data.StorageConfig.EnableNFS {
			assert.True(t, converter.EnableNFSVolume(data.StorageConfig), "not true")
		} else {
			assert.True(t, !converter.EnableNFSVolume(data.StorageConfig), "not true")
		}
		if data.StorageConfig.EnableLocal {
			assert.True(t, converter.EnableShareVolume(data.StorageConfig), "not true")
		} else {
			assert.True(t, !converter.EnableShareVolume(data.StorageConfig), "not true")
		}
	}
}
