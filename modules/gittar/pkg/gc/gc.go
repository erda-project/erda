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

package gc

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/gittar/conf"
	"github.com/erda-project/erda/modules/gittar/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/cron"
)

// all gittar repository root address
var repositoryRootAddr = "/repository"

// gc repository concurrent number
var concurrentNum = 0

// gc repository cron time
var cronExpression = ""

// cron task to clean repository
func ScheduledExecuteClean() {
	cronProcess := cron.New(
		cron.WithoutDLock(true),
	)

	cronExpression = conf.GitGCCronExpression()
	concurrentNum = conf.GitGCMaxNum()
	repositoryRootAddr = conf.RepoRoot()

	err := cronProcess.AddFunc(cronExpression, repositoryClean)
	if err != nil {
		panic(fmt.Errorf("cannot perform cleanup scheduled tasks, error: %v", err))
	}
	cronProcess.Start()
}

// query the project according to the root directory,
// then query each application according to the project,
// enter the application directory and execute 'git gc'
func repositoryClean() {
	logrus.Infof("gc: start gc")
	defer logrus.Infof("gc: end gc")
	projectFileInfos, err := ioutil.ReadDir(repositoryRootAddr)
	if err != nil {
		logrus.Errorf("gc: read dir %s error: %v", repositoryRootAddr, err)
		return
	}
	// initialize a waitGroup according to the number of concurrent
	var wait = limit_sync_group.NewSemaphore(concurrentNum)
	for _, projectFileInfo := range projectFileInfos {
		if !projectFileInfo.IsDir() {
			continue
		}
		var projectPath = repositoryRootAddr + "/" + projectFileInfo.Name()
		applicationFileInfos, err := ioutil.ReadDir(projectPath)
		if err != nil {
			logrus.Errorf("gc: read dir %s error: %v", projectPath, err)
			continue
		}
		for _, appFileInfo := range applicationFileInfos {
			if !appFileInfo.IsDir() {
				continue
			}
			// concurrently execute git gc commands
			wait.Add(1)
			go func(path string) {
				defer wait.Done()
				doGcCommand(path)
			}(projectPath + "/" + appFileInfo.Name())
		}
	}
	wait.Wait()
}

// execute the git gc command and print the returned information or error
func doGcCommand(path string) {
	// remove spaces to prevent injection attacks
	path = strings.Replace(path, " ", "", -1)
	cmd := exec.Command("sh", "-c", "cd '"+path+"' && git gc")
	output, err := cmd.CombinedOutput()
	logrus.Infof("gc: start gc path: %v", path)
	if err != nil {
		logrus.Errorf("gc: command run error: %v", err)
		return
	}
	logrus.Infof("gc: command output: %s", string(output))
	logrus.Infof("gc: end gc path: %v", path)
}
