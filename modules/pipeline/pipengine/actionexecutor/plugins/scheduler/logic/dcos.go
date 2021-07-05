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

package logic

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func splitTags(tagStr string) []string {
	return strings.FieldsFunc(tagStr, func(c rune) bool {
		return c == ','
	})
}

func BuildDcosConstraints(enable bool, labels map[string]string, preserveProjects map[string]struct{}, workspaceTags map[string]struct{}) [][]string {
	if !enable {
		return [][]string{}
	}
	matchTags := splitTags(labels[apistructs.LabelMatchTags])
	excludeTags := splitTags(labels[apistructs.LabelExcludeTags])
	var cs [][]string
	anyTagDisable := false
	if projectId, ok := labels["DICE_PROJECT"]; ok {
		_, exists := preserveProjects[projectId]
		if exists {
			anyTagDisable = true
			cs = append(cs, []string{"dice_tags", "LIKE", `.*\b` + apistructs.TagProjectPrefix + projectId + `\b.*`})
		} else {
			cs = append(cs, []string{"dice_tags", "UNLIKE", `.*\b` + apistructs.TagProjectPrefix + `[^,]+` + `\b.*`})
		}
	}

	if envTag, ok := labels["DICE_WORKSPACE"]; ok {
		_, exists := workspaceTags[envTag]
		if exists {
			cs = append(cs, []string{"dice_tags", "LIKE", `.*\b` + apistructs.TagWorkspacePrefix + envTag + `\b.*`})
			anyTagDisable = true
		} else {
			cs = append(cs, []string{"dice_tags", "UNLIKE", `.*\b` + apistructs.TagWorkspacePrefix + `[^,]+` + `\b.*`})
		}
	}
	if len(matchTags) == 0 && !anyTagDisable {
		cs = append(cs, []string{"dice_tags", "LIKE", `.*\bany\b.*`})
	} else if len(matchTags) != 0 && anyTagDisable {
		for _, matchTag := range matchTags {
			cs = append(cs, []string{"dice_tags", "LIKE", `.*\b` + matchTag + `\b.*`})
		}
	} else if len(matchTags) != 0 && !anyTagDisable {
		for _, matchTag := range matchTags {
			// The bigdata tag does not coexist with any
			if matchTag == "bigdata" {
				cs = append(cs, []string{"dice_tags", "LIKE", `.*\b` + matchTag + `\b.*`})
			} else {
				cs = append(cs, []string{"dice_tags", "LIKE", `.*\b` + apistructs.TagAny + `\b.*|.*\b` + matchTag + `\b.*`})
			}
		}
	}
	for _, excludeTag := range excludeTags {
		cs = append(cs, []string{"dice_tags", "UNLIKE", `.*\b` + excludeTag + `\b.*`})
	}
	return cs
}

// GetAndSetTokenAuth call this in goroutine
func GetAndSetTokenAuth(client *httpclient.HTTPClient, executorName string) {
	waitTime := 500 * time.Millisecond
	cnt := 0
	userNotSetAuthToken := 10
	for cnt < userNotSetAuthToken {
		select {
		case <-time.After(waitTime):
			if token, ok := os.LookupEnv("AUTH_TOKEN"); ok {
				if len(token) > 0 {
					client.TokenAuth(token)
					// Update every 2 hours, need to be less than getDCOSTokenAuthPeriodically medium period (24 hours)
					waitTime = 2 * time.Hour
					logrus.Debugf("executor %s got auth token, would re-get in %s later",
						executorName, waitTime.String())
				} else {
					if waitTime < 24*time.Hour {
						waitTime = waitTime * 2
					}
					logrus.Debugf("executor %s not got auth token, try again in %s later",
						executorName, waitTime.String())
				}
			} else {
				// The user has not set token auth, retry userNotSetAuthToken times
				cnt++
			}
		}
	}

	logrus.Debugf("env AUTH_TOKEN not set, executor(%s) goroutine exit", executorName)
}
