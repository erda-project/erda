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

func Test_EnablePipelineVolume(t *testing.T) {
	var tables = []struct {
		StorageConfig            StorageConfig
		EnablePipelineVolume     bool
		EnableCloseNetworkVolume bool
		EnableShareVolume        bool
	}{
		{
			StorageConfig: StorageConfig{
				EnableNFS:   false,
				EnableLocal: true,
			},
			EnablePipelineVolume:     false,
			EnableCloseNetworkVolume: false,
			EnableShareVolume:        true,
		},
		{
			StorageConfig: StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
			EnablePipelineVolume:     true,
			EnableCloseNetworkVolume: true,
			EnableShareVolume:        false,
		},
		{
			StorageConfig: StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
			EnablePipelineVolume:     true,
			EnableCloseNetworkVolume: true,
			EnableShareVolume:        false,
		},
	}
	for _, data := range tables {
		if data.EnablePipelineVolume {
			assert.True(t, data.StorageConfig.EnablePipelineVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnablePipelineVolume(), "not true")
		}
		if data.EnableCloseNetworkVolume {
			assert.True(t, data.StorageConfig.EnableNFSVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnableNFSVolume(), "not true")
		}
		if data.EnableShareVolume {
			assert.True(t, data.StorageConfig.EnableShareVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnableShareVolume(), "not true")
		}
	}
}
