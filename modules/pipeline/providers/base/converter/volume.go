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
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

// EnableNFSVolume
// whether to close the mounting of the network storage
// after closing, some special pipeline syntax ( ${{ dirs.xxx }} or old ${xxx} ) will not be available
func EnableNFSVolume(cfg *basepb.StorageConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.EnableNFS
}

// EnableShareVolume
// whether to open shared storage
// after open, the context directory in the pipeline will be shared
func EnableShareVolume(cfg *basepb.StorageConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.EnableLocal
}
