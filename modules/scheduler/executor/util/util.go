package util

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	// 检查依赖是否在service集合中
	for _, depends := range svcSet {
		for d := range depends {
			if _, ok := svcSet[d]; !ok {
				return nil, errors.Errorf("not found service: %s, it's in depends but not in runtime services", d)
			}
		}
	}

	handleDepends := func(svcDepMap map[string]serviceDepends) ([]string, error) {
		independents := make([]string, 0)

		// 找出没有依赖的 service
		for name, depends := range svcDepMap {
			if len(depends) == 0 {
				independents = append(independents, name)
			}
		}

		// 如果找不到一个无依赖的 service，说明存在死循环.
		if len(independents) == 0 {
			return nil, errors.New("dead loop in the service dependency")
		}

		// 清除这一批无依赖的 service
		for _, name := range independents {
			delete(svcDepMap, name)
		}

		// 清理依赖
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
		// 理出有向无环图中出度为0的节点, 并将这些节点从svcSet中删除
		independents, err := handleDepends(svcSet)
		if err != nil {
			return nil, err
		}
		sort.Strings(independents)

		// 处于同一层的service可以并行的去创建
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
			// bigdata标签不与any共存
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
					// 每2小时更新一次，需要小于getDCOSTokenAuthPeriodically中周期（24小时）
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
				// 用户未设置token auth, 重试 userNotSetAuthToken 次
				cnt++
			}
		}
	}

	logrus.Debugf("env AUTH_TOKEN not set, executor(%s) goroutine exit", executorName)
}
