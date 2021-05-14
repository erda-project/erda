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

package query

import (
	"encoding/json"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/modules/monitor/notify/template/db"
	"github.com/erda-project/erda/modules/monitor/notify/template/model"
	"github.com/erda-project/erda/modules/monitor/utils"
)

var (
	//notify template configuration file
	templateMap map[string]model.Model
)

func getAllNotifyTemplates() (list []model.Model) {
	for k := range templateMap {
		list = append(list, templateMap[k])
	}
	return
}

//obtain notify template list
func getNotifyTemplateList(scope, name, nType string) (list []*model.GetNotifyRes) {
	for _, v := range templateMap {
		if len(v.Metadata.Scope) > 0 {
			if scope != "" && v.Metadata.Scope[0] != scope {
				continue
			}
			if name != "" && v.Metadata.Name != name {
				continue
			}
			if nType != "" && v.Metadata.Type != nType {
				continue
			}
			m := model.GetNotifyRes{
				ID:   v.ID,
				Name: v.Metadata.Name,
			}
			list = append(list, &m)
		}
	}
	return
}

func ToNotifyConfig(c *model.CreateUserDefineNotifyTemplate) (*db.NotifyConfig, error) {
	//generate template_id
	templateID, err := utils.UUID()
	if err != nil {
		return nil, err
	}
	metadata := model.Metadata{
		Name:   c.Name,
		Type:   "custom",
		Module: "monitor",
		Scope:  []string{c.Scope},
	}
	metadataStr, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	behavior := model.Behavior{
		Group: c.Group,
	}
	behaviorStr, err := yaml.Marshal(behavior)
	if err != nil {
		return nil, err

	}
	templates := make([]model.Templates, 0)
	for i := range c.Trigger {
		trigger := strings.Split(c.Trigger[i], ",")
		targets := strings.Split(c.Targets[i], ",")
		template := model.Templates{
			Trigger: trigger,
			Targets: targets,
			I18n: []string{
				"zh-CN",
				"en-US",
			},
			Render: model.Render{},
		}
		if len(c.Formats)-1 >= i {
			template.Render.Formats = c.Formats[i]
		}
		if len(c.Title)-1 >= i {
			template.Render.Title = c.Title[i]
		}
		if len(c.Template)-1 >= i {
			template.Render.Template = c.Template[i]
		}
		templates = append(templates, template)
	}
	templateStr, err := yaml.Marshal(templates)
	if err != nil {
		return nil, err
	}
	customize := &db.NotifyConfig{
		NotifyID:  templateID,
		Metadata:  string(metadataStr),
		Behavior:  string(behaviorStr),
		Templates: string(templateStr),
		Scope:     c.Scope,
		ScopeID:   c.ScopeID,
	}
	return customize, nil
}

func ToNotify(u *model.UpdateNotifyReq) (*db.Notify, error) {
	var notify db.Notify
	notify.ID = uint(u.ID)
	notifyId, err := json.Marshal(u.TemplateId)
	if err != nil {
		return nil, err
	}
	notify.NotifyID = string(notifyId)
	t := model.Target{
		GroupID:  u.NotifyGroupID,
		Channels: u.Channels,
	}
	target, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	notify.Target = string(target)
	attribute, err := json.Marshal(u.Attribute)
	if err != nil {
		return nil, err
	}
	notify.Attributes = string(attribute)
	return &notify, nil
}

func ToNotifyRecord(n *model.NotifyRecord) *db.NotifyRecord {
	record := &db.NotifyRecord{}
	record.NotifyId = n.NotifyId
	record.NotifyName = n.NotifyName
	record.ScopeType = n.ScopeType
	record.ScopeId = n.ScopeId
	record.GroupId = n.GroupId
	record.NotifyGroup = n.NotifyGroup
	record.Title = n.Title
	record.NotifyTime = time.Unix(n.NotifyTime/1000, 0)
	record.CreateTime = time.Unix(n.CreateTime, 0)
	record.UpdateTime = time.Unix(n.UpdateTime, 0)
	return record
}
