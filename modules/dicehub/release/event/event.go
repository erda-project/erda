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

// Package event DiceHub important operation event sending logic
package event

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/release/db"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

const (
	// ReleaseEventCreate release create event type
	ReleaseEventCreate = "create"
	// ReleaseEventUpdate release update event type
	ReleaseEventUpdate = "update"
	// ReleaseEventDelete release delete event type
	ReleaseEventDelete = "delete"
)

// SendReleaseEvent send release event processing logic
func SendReleaseEvent(action string, release *db.Release) {
	content := &apistructs.ReleaseEventData{
		ReleaseID:     release.ReleaseID,
		ReleaseName:   release.ReleaseName,
		Addon:         release.Addon,
		Version:       release.Version,
		ClusterName:   release.ClusterName,
		OrgID:         release.OrgID,
		ProjectID:     release.ProjectID,
		ApplicationID: release.ApplicationID,
		UserID:        release.UserID,
		CrossCluster:  release.CrossCluster,
		CreatedAt:     release.CreatedAt,
		UpdatedAt:     release.UpdatedAt,
	}

	// Allow events to be lost without affecting the main process
	_, err := httpclient.New().Post("http://" + discover.EventBox()).
		Path("/api/dice/eventbox/message/create").
		JSONBody(apistructs.EventCreateRequest{
			EventHeader: apistructs.EventHeader{
				Event:     "release",
				Action:    action,
				OrgID:     strconv.FormatInt(content.OrgID, 10),
				ProjectID: strconv.FormatInt(content.ProjectID, 10),
				TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
			},
			Sender:  "dicehub",
			Content: content,
		}).Do().DiscardBody()

	if err != nil {
		logrus.Warnf("send release event fail, action: %s, release: %s, err: %v", action, release.ReleaseID, err)
	}
}
