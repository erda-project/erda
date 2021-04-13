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

package api

import (
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/gittar/models"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

type HookRequest struct {
	Id         int64  `json:"id"`
	Url        string `json:"url"`
	Name       string `json:"name"`
	PushEvents bool   `json:"push_events"`
}

// GetRepoStats function
func GetHooks(context *webcontext.Context) {
	repository := context.Repository

	hooks, err := context.Service.GetProjectHooks(repository)
	if err != nil {
		context.Abort(ERROR_DB)
		return
	}
	context.Success(hooks)
}

func GetHookDetail(context *webcontext.Context) {
	repository := context.Repository
	id, err := strconv.ParseInt(context.Param("id"), 10, 32)
	if err != nil {
		context.AbortWithStatus(400, ERROR_ARG_ID)
		return
	}
	hook := models.WebHook{}
	if context.DBClient.First(&hook, "id= ? AND repo_id = ?",
		id, repository.ID).RecordNotFound() {
		context.AbortWithStatus(404, ERROR_HOOK_NOT_FOUND)
		return
	}

	context.Success(hook)
}

func AddHook(context *webcontext.Context) {
	repository := context.Repository
	request := HookRequest{}
	err := context.BindJSON(&request)
	if err != nil {
		context.AbortWithStatus(400, errors.New("request body parse failed"))
		return
	}

	hook := models.WebHook{
		Url:        request.Url,
		RepoID:     repository.ID,
		Name:       request.Name,
		IsActive:   true,
		PushEvents: request.PushEvents,
	}

	currentHook, err := context.Service.AddProjectHook(&hook)
	if err != nil {
		context.AbortWithStatus(500, errors.New("db error"))
		logrus.Errorf("create project hook error %v", err)
	}

	context.Success(currentHook)
}

func UpdateHook(context *webcontext.Context) {
	repository := context.Repository
	id, err := strconv.ParseInt(context.Param("id"), 10, 32)
	if err != nil {
		context.AbortWithStatus(400, errors.New("id parse failed"))
		return
	}

	request := HookRequest{}
	err = context.BindJSON(&request)
	if err != nil {
		context.AbortWithStatus(400, errors.New("request body parse failed"))
		return
	}

	hook := models.WebHook{}

	if context.DBClient.First(&hook, "id= ? AND repo_id = ?",
		id, repository.ID).RecordNotFound() {
		context.AbortWithStatus(404, errors.New("hook not found"))
		return
	}

	hook.PushEvents = request.PushEvents
	hook.Url = request.Url

	context.DBClient.Save(&hook)
	context.Success(hook)

}

func DeleteHook(context *webcontext.Context) {
	repository := context.Repository
	id, err := strconv.ParseInt(context.Param("id"), 10, 32)
	if err != nil {
		context.AbortWithStatus(400, errors.New("id parse failed"))
		return
	}
	hook := models.WebHook{}

	if context.DBClient.First(&hook, "id= ? AND repo_id = ?",
		id, repository.ID).RecordNotFound() {
		context.AbortWithStatus(404, errors.New("hook not found"))
		return
	}

	context.DBClient.Delete(hook)
	context.Success(nil)
}

func RequeueTask(context *webcontext.Context) {
	hookTask := models.WebHookTask{}
	id, err := strconv.ParseInt(context.Param("id"), 10, 32)
	if err != nil {
		context.AbortWithStatus(400)
		return
	}
	if context.DBClient.First(&hookTask, "id= ? ", id).RecordNotFound() {
		context.AbortWithStatus(404, errors.New("task not found"))
		return
	}
	models.AddToHookTaskQueue(&hookTask)
	context.Success(nil)
}

func AddSystemHook(context *webcontext.Context) {
	request := HookRequest{}
	err := context.BindJSON(&request)
	if err != nil {
		context.AbortWithStatus(400, errors.New("request body parse failed"))
		return
	}

	hook := models.WebHook{
		Url:        request.Url,
		Name:       request.Name,
		IsActive:   true,
		PushEvents: request.PushEvents,
	}
	result, err := context.Service.AddSystemHook(&hook)
	if err != nil {
		context.AbortWithStatus(500, errors.New("db error"))
		logrus.Errorf("create system hook error %v", err)
		return
	}
	context.Success(result)
}
