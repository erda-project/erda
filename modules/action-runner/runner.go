//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package actionrunner

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/template"
)

// Runner .
type Runner struct {
	Conf  *Conf
	queue chan *Task
	tasks int32
}

// New .
func New(Conf *Conf) *Runner {
	return &Runner{
		Conf:  Conf,
		queue: make(chan *Task, Conf.MaxTask),
	}
}

// Run .
func (r *Runner) Run() error {
	err := r.runStartUpCommands()
	if err != nil {
		return err
	}
	for i := 0; i < r.Conf.MaxTask; i++ {
		go r.worker()
	}
	go r.cleanBuildDir()
	r.reloadTasks()
	return nil
}

func (r *Runner) runStartUpCommands() error {
	for _, commandTemplate := range r.Conf.StartupCommands {
		commandStr := template.Render(commandTemplate, r.Conf.Params)
		cmd := exec.Command("/bin/bash", "-c", commandStr)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) cleanBuildDir() {
	interval := 15 * time.Minute
	for {
		fileInfoList, err := ioutil.ReadDir(r.Conf.BuildPath)
		if err != nil {
			logrus.Errorf("failed to read build dir %s err:%s", r.Conf.BuildPath, err)
		} else {
			for _, info := range fileInfoList {
				if info.IsDir() && strings.HasPrefix(info.Name(), "pipeline-task") {
					if info.ModTime().Add(time.Hour*time.Duration(r.Conf.FailedTaskKeepHours)).Unix() < time.Now().Unix() {
						buildDir := path.Join(r.Conf.BuildPath, info.Name())
						logrus.Infof("remove build dir %s", buildDir)
						os.RemoveAll(buildDir)
					}
				}
			}
		}
		time.Sleep(interval)
	}
}

func (r *Runner) reloadTasks() {
	interval := 20 * time.Second
	for {
		tasks := atomic.LoadInt32(&r.tasks)
		if tasks < int32(r.Conf.MaxTask) {
			list := r.fetchTasks()
			for _, task := range list {
				r.queue <- task
			}
		}
		time.Sleep(interval)
		if tasks > 0 {
			logrus.Infof("running task %d", tasks)
		}
	}
}

func (r *Runner) worker() {
	log := r.newLogger()
	for task := range r.queue {
		atomic.AddInt32(&r.tasks, 1)
		w := &worker{
			Task:        task,
			r:           r,
			conf:        r.Conf,
			contextPath: path.Join(r.Conf.BuildPath, task.JobID),
			log:         newJobLogger(log, task.JobID, task.Token),
		}
		w.Execute()
		atomic.AddInt32(&r.tasks, -1)
	}
}
