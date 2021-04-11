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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/conf"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type EventCallback struct {
	Name    string
	URLPath string
	Events  []string
}

var eventCallbacks = []EventCallback{qaGitMRCreateCallback}

var qaGitMRCreateCallback = EventCallback{
	Name:    "qa_git_mr_create",
	URLPath: "/api/callbacks/git-mr-create",
	Events:  []string{"git_create_mr"},
}

func (e *Endpoints) RegisterWebhooks() error {
	for _, callback := range eventCallbacks {
		hook := apistructs.CreateHookRequest{
			Name:   callback.Name,
			Events: callback.Events,
			URL:    strutil.Concat("http://", conf.SelfAddr(), callback.URLPath),
			Active: true,
			HookLocation: apistructs.HookLocation{
				Org:         "-1",
				Project:     "-1",
				Application: "-1",
			},
		}
		if err := e.bdl.CreateWebhook(hook); err != nil {
			return fmt.Errorf("failed to register hook %s to eventbox, err: %v, detail: %v", callback.Name, err, hook)
		}
		logrus.Infof("register hook %s to eventbox, detail: %v", callback.Name, hook)
	}
	return nil
}

func (e *Endpoints) GittarMRCreateCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	logrus.Infof("received mr request")
	if r.ContentLength == 0 {
		return apierrors.ErrDoGitMrCreateCallback.MissingParameter(apierrors.MissingRequestBody).ToResp(), nil
	}
	var mrCreateEvent apistructs.RepoCreateMrEvent
	if err := json.NewDecoder(r.Body).Decode(&mrCreateEvent); err != nil {
		return apierrors.ErrDoGitMrCreateCallback.InvalidParameter(err).ToResp(), nil
	}

	// 打印
	logrus.Infof("received mr event: %+v", mrCreateEvent)

	go func() {
		pipelineID, err := e.cq.TriggerByMR(mrCreateEvent.Content)
		if err != nil {
			logrus.Errorf("MR CQ failed, err: %v", err)
		} else {
			logrus.Infof("MR CQ triggered, pipelineID: %d", pipelineID)
		}
	}()

	return httpserver.OkResp(nil)
}
