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

package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type serviceDepends map[string]struct{}

func ParseServiceDependency(runtime *apistructs.ServiceGroup) ([][]*apistructs.Service, error) {
	if runtime == nil {
		return nil, errors.New("runtime could not be nil")
	}
	services := runtime.Services
	serviceMap := make(map[string]*apistructs.Service)

	svcSet := make(map[string]serviceDepends, len(services))
	for i, svc := range services {
		if _, ok := svcSet[svc.Name]; ok {
			return nil, errors.New("duplicate service name in runtime")
		}
		depends := serviceDepends{}
		for _, d := range svc.Depends {
			depends[d] = struct{}{}
		}
		svcSet[svc.Name] = depends
		serviceMap[svc.Name] = &services[i]
	}

	// Check if the dependency is in the service collection
	for _, depends := range svcSet {
		for d := range depends {
			if _, ok := svcSet[d]; !ok {
				return nil, errors.Errorf("not found service: %s, it's in depends but not in runtime services", d)
			}
		}
	}

	handleDepends := func(svcDepMap map[string]serviceDepends) ([]string, error) {
		independents := make([]string, 0)

		// Find out the service that has no dependencies
		for name, depends := range svcDepMap {
			if len(depends) == 0 {
				independents = append(independents, name)
			}
		}

		// If you can't find a non-dependent service, there is an infinite loop.
		if len(independents) == 0 {
			return nil, errors.New("dead loop in the service dependency")
		}

		// Clear this batch of non-dependent services
		for _, name := range independents {
			delete(svcDepMap, name)
		}

		// Cleanup dependencies
		for _, name := range independents {
			for job, depends := range svcDepMap {
				delete(depends, name)
				svcDepMap[job] = depends
			}
		}

		return independents, nil
	}

	layers := make([][]*apistructs.Service, 0, len(services))
	for len(svcSet) != 0 {
		// Sort out the nodes with degree 0 in the directed acyclic graph, and delete these nodes from svcSet
		independents, err := handleDepends(svcSet)
		if err != nil {
			return nil, err
		}
		sort.Strings(independents)

		// Services in the same layer can be created in parallel
		svcLayer := make([]*apistructs.Service, 0, len(independents))
		for _, name := range independents {
			if svcAddr, ok := serviceMap[name]; ok {
				svcLayer = append(svcLayer, svcAddr)
			}
		}
		if len(svcLayer) != 0 {
			layers = append(layers, svcLayer)
		}
	}

	return layers, nil
}

func ParseEnableTagOption(options map[string]string, key string, defaultValue bool) (bool, error) {
	enableTagStr, ok := options[key]
	if !ok {
		return defaultValue, nil
	}
	enableTag, err := strconv.ParseBool(enableTagStr)
	if err != nil {
		return false, err
	}
	return enableTag, nil
}

func ParsePreserveProjects(options map[string]string, key string) map[string]struct{} {
	ret := make(map[string]struct{})
	projectsStr, ok := options[key]
	if !ok {
		return map[string]struct{}{}
	}
	projects := splitTags(projectsStr)
	for _, p := range projects {
		ret[p] = struct{}{}
	}
	return ret
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

func CombineLabels(parent, child map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range parent {
		ret[k] = v
	}
	for k, v := range child {
		ret[k] = v
	}
	return ret
}

func splitTags(tagStr string) []string {
	return strings.FieldsFunc(tagStr, func(c rune) bool {
		return c == ','
	})
}

// call this in goroutine
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

func IsNotFound(err error) bool {
	if strings.Contains(err.Error(), "not found") {
		return true
	}
	return false
}

// GetClient get http client with cluster info.
func GetClient(clusterName string, manageConfig *apistructs.ManageConfig) (string, *httpclient.HTTPClient, error) {
	if manageConfig == nil {
		return "", nil, fmt.Errorf("cluster %s manage config is nil", clusterName)
	}

	inetPortal := "inet://"

	hcOptions := []httpclient.OpOption{
		httpclient.WithHTTPS(),
	}

	// check mange config type
	switch manageConfig.Type {
	case apistructs.ManageProxy, apistructs.ManageToken:
		// cluster-agent -> (register) cluster-dialer -> (patch) cluster-manager
		// -> (update) eventBox -> (update) scheduler -> scheduler reload executor
		if manageConfig.Token == "" || manageConfig.Address == "" {
			return "", nil, fmt.Errorf("token or address is empty")
		}

		hc := httpclient.New(hcOptions...)
		hc.BearerTokenAuth(manageConfig.Token)

		if manageConfig.Type == apistructs.ManageToken {
			return manageConfig.Address, hc, nil
		}

		// parseInetAddr parse inet addr, will add proxy header in custom http request
		return fmt.Sprintf("%s%s/%s", inetPortal, clusterName, manageConfig.Address), hc, nil
	case apistructs.ManageCert:
		if len(manageConfig.KeyData) == 0 ||
			len(manageConfig.CertData) == 0 {
			return "", nil, fmt.Errorf("cert or key is empty")
		}

		certBase64, err := base64.StdEncoding.DecodeString(manageConfig.CertData)
		if err != nil {
			return "", nil, err
		}
		keyBase64, err := base64.StdEncoding.DecodeString(manageConfig.KeyData)
		if err != nil {
			return "", nil, err
		}

		var certOption httpclient.OpOption

		certOption = httpclient.WithHttpsCertFromJSON(certBase64, keyBase64, nil)

		if len(manageConfig.CaData) != 0 {
			caBase64, err := base64.StdEncoding.DecodeString(manageConfig.CaData)
			if err != nil {
				return "", nil, err
			}
			certOption = httpclient.WithHttpsCertFromJSON(certBase64, keyBase64, caBase64)
		}
		hcOptions = append(hcOptions, certOption)

		return manageConfig.Address, httpclient.New(hcOptions...), nil
	default:
		return "", nil, fmt.Errorf("manage type is not support")
	}
}
