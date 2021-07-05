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

package sms

import (
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/template"

	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type MBoxSubscriber struct {
	bundle *bundle.Bundle
}

type MBoxData struct {
	Template      string            `json:"template"`
	Params        map[string]string `json:"params"`
	OrgID         int64             `json:"orgID"`
	Label         string            `json:"label"`
	DeduplicateID string            `json:"deduplicateId"`
}

type Option func(*MBoxSubscriber)

func New(bundle *bundle.Bundle) subscriber.Subscriber {
	return &MBoxSubscriber{
		bundle: bundle,
	}
}

func (d *MBoxSubscriber) Publish(dest string, content string, time int64, msg *types.Message) []error {
	errs := []error{}
	var userIDs []string
	err := json.Unmarshal([]byte(dest), &userIDs)
	if err != nil {
		return []error{err}
	}
	var mboxData MBoxData
	err = json.Unmarshal([]byte(content), &mboxData)
	if err != nil {
		return []error{err}
	}
	title, ok := mboxData.Params["title"]
	if !ok {
		title = "站内信通知"
	}
	err = d.bundle.CreateMBox(&apistructs.CreateMBoxRequest{
		Title:         template.Render(title, mboxData.Params),
		Content:       template.Render(mboxData.Template, mboxData.Params),
		OrgID:         mboxData.OrgID,
		UserIDs:       userIDs,
		Label:         mboxData.Label,
		DeduplicateID: mboxData.DeduplicateID,
	})
	if err != nil {
		return []error{err}
	}
	return errs
}

func (d *MBoxSubscriber) Status() interface{} {
	return nil
}

func (d *MBoxSubscriber) Name() string {
	return "MBOX"
}
