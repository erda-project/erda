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

package endpoints

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	rulepb "github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/apistructs"
)

type EventInfo struct {
	apistructs.EventHeader
	Scope   string
	ScopeID string
}

func (e *Endpoints) FireRule(ctx context.Context, content interface{}, eventInfo EventInfo) error {
	marshal, err := json.Marshal(content)
	if err != nil {
		return err
	}
	params := make(map[string]interface{})
	if err := json.Unmarshal(marshal, &params); err != nil {
		return err
	}

	config, err := e.SetConfigInfo(eventInfo)
	if err != nil {
		return err
	}
	params["config"] = config
	env, err := structpb.NewValue(params)
	if err != nil {
		return err
	}

	eventType := eventInfo.Event
	res, err := e.ruleExecutor.Fire(ctx, &rulepb.FireRequest{
		Scope:     eventInfo.Scope,
		ScopeID:   eventInfo.ScopeID,
		EventType: eventType,
		Env: map[string]*structpb.Value{
			eventType: env,
		},
	})
	if err != nil {
		return err
	}
	logrus.Infof("%s rule executor result %v", eventType, res)
	return nil
}

func (e *Endpoints) SetConfigInfo(eventInfo EventInfo) (map[string]interface{}, error) {
	logrus.Infof("set config eventInfo %v", eventInfo)
	res := make(map[string]interface{})
	res["eventType"] = eventInfo.Event
	projectID, _ := strconv.Atoi(eventInfo.ProjectID)
	project, err := e.bdl.GetProject(uint64(projectID))
	if err != nil {
		return nil, err
	}
	if project != nil {
		res["project"] = map[string]interface{}{
			"id":          project.ID,
			"name":        project.Name,
			"displayName": project.DisplayName,
		}
	}

	appID, _ := strconv.Atoi(eventInfo.ApplicationID)
	if appID <= 0 {
		return res, nil
	}
	app, err := e.bdl.GetApp(uint64(appID))
	if err != nil {
		return nil, err
	}
	if app != nil {
		res["app"] = map[string]interface{}{
			"id":          app.ID,
			"name":        app.Name,
			"displayName": app.DisplayName,
		}
	}
	return res, nil
}
