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

package autoop

import (
	"context"
	"path/filepath"
	"time"

	"github.com/robfig/cron"

	"github.com/erda-project/erda/apistructs"
)

const (
	basePath = "/var/lib/autoop"
)

var (
	scriptsPath    = filepath.Join(basePath, "scripts")
	newScriptsPath = filepath.Join(basePath, "new-scripts")
	logsPath       = filepath.Join(basePath, "logs")
)

// Action Automation operation script related information and status
type Action struct {
	Path        string
	Name        string
	Md5         string
	ClusterInfo *apistructs.ClusterInfo
	Running     bool
	RunTime     time.Time
	Env         map[string]string
	CronTime    string
	Nodes       string
	Cron        *cron.Cron
	Context     context.Context
	CancelFunc  context.CancelFunc
}

// NewAction Returns an Action according to the automation operation script path
func NewAction(path string) *Action {
	return &Action{Path: filepath.Dir(path), Name: filepath.Base(path)}
}

// File Returns the file name of automatic operation file
func (a *Action) File(name string) string {
	return filepath.Join(a.Path, a.Name, name)
}

// UpdateTaskInfo Send the running status of the automated operation and maintenance script
//func (a *Action) UpdateTaskInfo(running, lastStatus bool, lastError string) error {
//	now := time.Now()
//	ms := now.Unix()*1e3 + int64(now.Nanosecond())/1e6
//	taskInfo := apistructs.TaskInfo{
//		ScriptName:  a.Name,
//		ClusterName: clusterInfo.Name,
//		Md5:         a.Md5,
//	}
//	if running {
//		taskInfo.Running = running
//		taskInfo.StartAt = ms
//	} else {
//		taskInfo.LastStatus = lastStatus
//		taskInfo.EndAt = ms
//		if !lastStatus {
//			taskInfo.LastError = lastError
//			taskInfo.ErrorAt = ms
//		}
//	}
//	token, err := getToken()
//	if err != nil {
//		return err
//	}
//	var v apistructs.Header
//	res, err := httpclient.New().Post(settings.OpenAPIURL).
//		Path("/api/script/task").Header("Authorization", token).JSONBody(taskInfo).Do().JSON(&v)
//	if err != nil {
//		return fmt.Errorf("update task info failed: %s", err.Error())
//	}
//	resetToken(res.StatusCode())
//	if !res.IsOK() {
//		return fmt.Errorf("update task info failed: status code is %d", res.StatusCode())
//	}
//	if !v.Success {
//		return fmt.Errorf("update task info failed: %s %s", v.Error.Code, v.Error.Msg)
//	}
//	return nil
//}
