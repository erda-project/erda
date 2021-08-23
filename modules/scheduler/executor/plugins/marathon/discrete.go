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

package marathon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/loop"
)

var (
	errAppInstanceNum         = errors.New("exists app with instance > 1 when HOST_UNIQUE = true")
	errNoNeedBuildPlaceHolder = errors.New("no need to build placeholder")
	errDiffConstraints        = errors.New("not every app has same constrains")

	errCheckAppsTimeout = errors.New("check all apps running failed: tried 10 times with 5s interval")
)

// placeHolderHosts
// key: appid
// value: host ip
type placeHolderHosts map[string]string

func hostConstraint(host string) Constraint {
	return Constraint{"hostname", "LIKE", host}
}

// TODO: Method attributable to App struct
func updateAppConstraints(app *App, constraint Constraint) {
	app.Constraints = append(app.Constraints, constraint)
}

// TODO: Rename the function name, it should belong to the method of Group struct
func updateGroupByPlaceHolderHosts(g *Group, pchosts placeHolderHosts) error {
	for i := range g.Apps {
		host, ok := pchosts[g.Apps[i].Id]
		if !ok {
			errmsg := fmt.Sprintf("[alert][BUG] updateGroupByPlaceHolderHosts: pchosts: %+v, g: %+v", pchosts, g)
			logrus.Error(errmsg)
			return errors.New(errmsg)
		}
		updateAppConstraints(&g.Apps[i], hostConstraint(host))
	}
	return nil
}

func (m *Marathon) createPlaceHolderGroup(sg *apistructs.ServiceGroup, origin *Group) (placeHolderHosts, error) {
	newgroup, exactappMap, err := buildPlaceHolderGroup(sg, origin)
	if err != nil {
		return nil, err
	}
	logrus.Infof("createPlaceHolderGroup: placeholder: %+v\n", newgroup) // debug print

	if _, err := m.putGroup(http.MethodPost, newgroup, true); err != nil {
		return nil, err
	}
	pcgroup, err := m.makeSureAppsRunning(sg)
	if err != nil {
		return nil, err
	}
	pchosts := placeHolderHosts{}
	for i := range pcgroup.Apps {
		exactapps := exactappMap[pcgroup.Apps[i].Id]
		if len(pcgroup.Apps[i].Tasks) != len(exactapps) {
			errmsg := fmt.Sprintf("[alert][BUG] createPlaceHolderGroup: pcgroup: %v, exactapps: %v", pcgroup, exactapps)
			logrus.Error(errmsg)
			return nil, errors.New(errmsg)
		}
		for j, appid := range exactapps {
			pchosts[appid] = pcgroup.Apps[i].Tasks[j].Host
		}
	}
	return pchosts, nil
}

func (m *Marathon) makeSureAppsRunning(sg *apistructs.ServiceGroup) (*Group, error) {
	var allAppsRunning bool
	group := Group{}
	if err := loop.New(loop.WithMaxTimes(10), loop.WithInterval(5*time.Second)).Do(func() (bool, error) {
		g, err := m.getGroupWithDefaultParam(m.buildGroupID(sg))
		if err != nil {
			return false, err
		}
		for i := range g.Apps {
			if g.Apps[i].TasksRunning == g.Apps[i].Instances {
				allAppsRunning = true
				group = *g
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return nil, err
	}
	if !allAppsRunning {
		return nil, errCheckAppsTimeout
	}
	return &group, nil
}

// buildPlaceHolderGroup According to the resources required by the service in `runtime`,
// Construct a servicegroup to occupy resources
// return (placeholderGroup, map[placeholder-appid][]exact-appid, err)
// NOTE:
// 1. For apps with instances> 1 in the app list, directly return errAppInstanceNum
// such as： {
//   app1: {
//     instances: N (N>1),
//   },
//   app2: {
//     instances: 1,
//   },
// }
// 2. If there is only `no more than` 1 app in the app list, return errNoNeedBuildPlaceHolder
// 3. If the constrains of each app in the app list are not all the same, return errDiffConstraints
func buildPlaceHolderGroup(sg *apistructs.ServiceGroup, originGroup *Group) (*Group, map[string][]string, error) {
	appNum := len(originGroup.Apps)
	if appNum <= 1 {
		return nil, nil, errNoNeedBuildPlaceHolder
	}
	for _, app := range originGroup.Apps {
		if app.Instances > 1 {
			return nil, nil, errAppInstanceNum
		}
		if !isSameConstraints(originGroup.Apps[0].Constraints, app.Constraints) {
			return nil, nil, errDiffConstraints
		}
	}

	newgroup, err := deepcopyGroup(originGroup)
	if err != nil {
		return nil, nil, err
	}
	appgroups, err := classifyApps(sg.ScheduleInfo.HostUniqueInfo, &newgroup)
	if err != nil {
		return nil, nil, err
	}
	r := newgroup
	r.Apps = []App{}
	exactappMap := make(map[string][]string)
	for _, apps := range appgroups {
		for _, app := range apps {
			exactappMap[apps[0].Id] = append(exactappMap[apps[0].Id], app.Id)
		}
		r.Apps = append(r.Apps, apps[0])
		thisone := len(r.Apps) - 1
		r.Apps[thisone].Container.Docker.Image = conf.PlaceHolderImage()
		r.Apps[thisone].Cmd = "sh -c \"while true; do sleep 5; done\""
		r.Apps[thisone].HealthChecks = []AppHealthCheck{
			{
				Protocol:               "COMMAND",
				Command:                &AppHealthCheckCommand{Value: "echo 1"},
				MaxConsecutiveFailures: apistructs.HealthCheckDuration / 15,
			},
		}
		r.Apps[thisone].Args = []string{}
		r.Apps[thisone].Constraints = append(r.Apps[thisone].Constraints,
			Constraint([]string{"hostname", "UNIQUE"}))
		r.Apps[thisone].Instances = len(apps)
		// Set volumes to nil to prevent the influence of localpv in the original group
		r.Apps[thisone].Container.Volumes = nil
	}
	return &r, exactappMap, nil
}

func classifyApps(groups [][]string, g *Group) ([][]App, error) {
	appgroups := make([][]App, len(groups))
	for _, group := range groups {
		if !sort.StringsAreSorted(group) {
			sort.Strings(group)
		}
	}

	for i := range g.Apps {
		splited := strings.Split(g.Apps[i].Id, "/")
		servicename := splited[len(splited)-1]
		var found bool
		for j, group := range groups {
			if idx := sort.SearchStrings(group, servicename); idx < len(group) &&
				group[idx] == servicename {
				appgroups[j] = append(appgroups[j], g.Apps[i])
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("failed to classifyApps, not found service: %v in groups: %v",
				servicename, groups)
		}
	}
	logrus.Infof("classifyApps: appgroups: %+v\n", appgroups) // debug print
	return appgroups, nil
}

func isSameConstraints(c1, c2 []Constraint) bool {
	if len(c1) != len(c2) {
		return false
	}
	for _, c1elem := range c1 {
		equal := false
		for _, c2elem := range c2 {
			if c1elem.Equal(c2elem) {
				equal = true
				break
			}
		}
		if !equal {
			return false
		}
	}
	return true
}

func deepcopyGroup(origin *Group) (Group, error) {
	marshalled, err := json.Marshal(origin)
	if err != nil {
		return Group{}, err
	}
	r := Group{}
	if err := json.Unmarshal(marshalled, &r); err != nil {
		return Group{}, err
	}
	return r, nil
}
