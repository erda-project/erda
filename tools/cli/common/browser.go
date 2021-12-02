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
	"strconv"
	"strings"

	"github.com/pkg/browser"

	"github.com/erda-project/erda/tools/cli/command"
)

type ErdaEntity string

var (
	OrgEntity     ErdaEntity = "org"
	ProjectEntity ErdaEntity = "project"
	AppEntity     ErdaEntity = "app"
)

func Open(ctx *command.Context, entity ErdaEntity, org string, orgId, projectId, applicationId uint64) error {
	if org == "" && orgId != 0 {
		o, err := GetOrgDetail(ctx, strconv.FormatUint(orgId, 10))
		if err != nil {
			return err
		}
		org = o.Name
	}

	paths := []string{strings.Replace(ctx.CurrentOpenApiHost, "openapi.", "", 1)}
	switch entity {
	case OrgEntity:
		paths = append(paths, org)
	case ProjectEntity:
		paths = append(paths, org, "dop/projects", strconv.FormatUint(projectId, 10))
	case AppEntity:
		paths = append(paths, org, "dop/projects", strconv.FormatUint(projectId, 10),
			"apps", strconv.FormatUint(applicationId, 10))
	}
	url := strings.Join(paths, "/")
	err := browser.OpenURL(url)
	if err != nil {
		return err
	}

	return nil
}
