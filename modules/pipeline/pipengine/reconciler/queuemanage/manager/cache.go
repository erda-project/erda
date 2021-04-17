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

package manager

import (
	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (mgr *defaultManager) GetPipelineCaches() map[uint64]*spec.Pipeline {
	mgr.pCacheLock.Lock()
	defer mgr.pCacheLock.Unlock()

	copied := deepcopy.Copy(mgr.pipelineCaches)
	r, ok := copied.(map[uint64]*spec.Pipeline)
	if !ok {
		return make(map[uint64]*spec.Pipeline)
	}
	return r
}
