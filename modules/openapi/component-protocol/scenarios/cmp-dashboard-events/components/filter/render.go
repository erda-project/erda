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

package filter

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &ComponentFilter{}
}

func (f *ComponentFilter) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	_ apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("context bundle can not be empty")
	}
	f.ctxBdl = bdl

	if err := f.GenComponentState(component); err != nil {
		return err
	}

	if err := f.SetComponentValue(); err != nil {
		return err
	}
	return nil
}

func (f *ComponentFilter) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	f.State = state
	return nil
}

func (f *ComponentFilter) SetComponentValue() error {
	userID := f.ctxBdl.Identity.UserID
	orgID := f.ctxBdl.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SNamespace,
		ClusterName: f.State.ClusterName,
	}

	data, err := f.ctxBdl.Bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list := data.Slice("data")

	devNs := Option{
		Label: "dev",
		Value: "dev",
	}
	testNs := Option{
		Label: "test",
		Value: "test",
	}
	stagingNs := Option{
		Label: "staging",
		Value: "staging",
	}
	productionNs := Option{
		Label: "production",
		Value: "production",
	}
	addonNs := Option{
		Label: "addons",
		Value: "addons",
	}
	pipelineNs := Option{
		Label: "pipelines",
		Value: "pipelines",
	}
	defaultNs := Option{
		Label: "default",
		Value: "default",
	}
	systemNs := Option{
		Label: "system",
		Value: "system",
	}
	otherNs := Option{
		Label: "others",
		Value: "others",
	}

	for _, obj := range list {
		name := obj.String("metadata", "name")
		option := Option{
			Label: name,
			Value: name,
		}
		if suf, ok := hasSuffix(name); ok && strings.HasPrefix(name, "project-") {
			displayName, err := f.getDisplayName(name)
			if err == nil {
				option.Label = displayName
				switch suf {
				case "-dev":
					devNs.Children = append(devNs.Children, option)
				case "-test":
					testNs.Children = append(testNs.Children, option)
				case "-prod":
					productionNs.Children = append(productionNs.Children, option)
				case "-staging":
					stagingNs.Children = append(stagingNs.Children, option)
				}
				continue
			}
		}
		if strings.HasPrefix(name, "addon-") || strings.HasPrefix(name, "group-addon-") {
			addonNs.Children = append(addonNs.Children, option)
			continue
		}
		if strings.HasPrefix(name, "pipeline-") {
			pipelineNs.Children = append(pipelineNs.Children, option)
			continue
		}
		if name == "default" {
			defaultNs.Children = append(defaultNs.Children, option)
			continue
		}
		if name == "kube-system" || name == "erda-system" {
			systemNs.Children = append(systemNs.Children, option)
			continue
		}
		otherNs.Children = append(otherNs.Children, option)
	}

	f.State.Conditions = nil
	namespaceCond := Condition{
		Key:   "namespace",
		Label: "Namespace",
		Type:  "select",
		Fixed: true,
	}
	for _, option := range []Option{devNs, testNs, productionNs, stagingNs, addonNs, pipelineNs, defaultNs, systemNs, otherNs} {
		if option.Children != nil {
			sort.Slice(option.Children, func(i, j int) bool {
				return option.Children[i].Label < option.Children[j].Label
			})
			namespaceCond.Options = append(namespaceCond.Options, option)
		}
	}
	f.State.Conditions = append(f.State.Conditions, namespaceCond)

	f.State.Conditions = append(f.State.Conditions, Condition{
		Key:   "type",
		Label: "Event Type",
		Type:  "select",
		Fixed: true,
		Options: []Option{
			{
				Label: "Normal",
				Value: "Normal",
			},
			{
				Label: "Warning",
				Value: "Warning",
			},
		},
	})

	f.Operations = make(map[string]interface{})
	f.Operations["filter"] = Operation{
		Key:    "filter",
		Reload: true,
	}
	return nil
}

func (f *ComponentFilter) getDisplayName(name string) (string, error) {
	splits := strings.Split(name, "-")
	if len(splits) != 3 {
		return "", errors.New("invalid name")
	}
	id := splits[1]
	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}
	project, err := f.ctxBdl.Bdl.GetProject(uint64(num))
	if err != nil {
		return "", err
	}
	return project.DisplayName, nil
}

func hasSuffix(name string) (string, bool) {
	suffixes := []string{"-dev", "-staging", "-test", "-prod"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return suffix, true
		}
	}
	return "", false
}
