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

package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RELEASEOPEN = command.Command{
	Name:       "open",
	ParentName: "RELEASE",
	ShortHelp:  "Open the release page in browser",
	Example:    "$ erda-cli release open --org=<name> --release-id=<id>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "release-id", Doc: "the id of a release", DefaultValue: ""},
	},
	Run: ReleaseOpen,
}

func ReleaseOpen(ctx *command.Context, orgId uint64, org, release string) error {
	checkOrgParam(org, orgId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	r, err := common.GetReleaseDetail(ctx, orgId, release)
	if err != nil {
		return err
	}

	err = common.Open2(ctx, org, orgId, func() []string {
		params := url.Values{}
		params.Add("applicationId", strconv.FormatInt(r.ApplicationID, 10))
		params.Add("q", release)
		return []string{
			fmt.Sprintf("dop/projects/%d/apps/%d/release?", r.ProjectID, r.ApplicationID),
			params.Encode(),
		}
	})
	if err != nil {
		return err
	}

	return nil
}
