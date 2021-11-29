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

package common

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func GetSerivceList(ctx *command.Context, orgId, applicationId uint64, workspace, runtime string) (
	map[string]*apistructs.RuntimeInspectServiceDTO, error) {
	r, err := GetRuntimeDetail(ctx, orgId, applicationId, workspace, runtime)
	if err != nil {
		return nil, err
	}

	return r.Data.Services, err
}
