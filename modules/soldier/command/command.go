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

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/modules/soldier/settings"
)

const (
	cmdDocker = "docker"
)

type ActionDockerLogs struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Container string `json:"container"`
	Since     string `json:"since"`
	Tail      string `json:"tail"`
}

func (a Action) DockerLogs() ([]string, error) {
	if a.Args == nil {
		return nil, errors.New("args required")
	}
	v := ActionDockerLogs{
		Port: 2375,
	}
	if err := json.Unmarshal(*a.Args, &v); err != nil {
		return nil, err
	}
	if v.Host == "" {
		return nil, errors.New("host required")
	}
	if !isPort(v.Port) {
		return nil, errors.New("port invalid")
	}
	if v.Container == "" {
		return nil, errors.New("container required")
	}
	args := make([]string, 0, 6)
	if settings.ForwardPort >= 0 {
		args = append(args, "-H", fmt.Sprintf("tcp://%s:%d", v.Host, v.Port))
	}
	args = append(args, "logs")
	if v.Since == "" && v.Tail == "" {
		args = append(args, "--tail", "100")
	} else {
		if v.Since != "" {
			args = append(args, "--since", v.Since)
		}
		if v.Tail != "" {
			args = append(args, "--tail", v.Tail)
		}
	}
	args = append(args, v.Container)
	return args, nil
}

type ActionCleanCI struct {
	Days int `json:"days"`
}

func (a Action) CleanCI() ([]string, error) {
	if a.Args == nil {
		return nil, errors.New("args required")
	}
	v := ActionCleanCI{
		Days: 7,
	}
	if err := json.Unmarshal(*a.Args, &v); err != nil {
		return nil, err
	}
	if v.Days < 1 || v.Days > 10000 {
		return nil, errors.New("days invalid")
	}
	return []string{"/netdata/devops/ci/pipelines",
		"-type", "d",
		"-maxdepth", "1",
		"-mtime", fmt.Sprintf("+%d", v.Days),
		"-exec", "rm", "-rf", "{}", ";",
	}, nil
}

func (a Action) CommandContext(ctx context.Context) (cmd *exec.Cmd, err error) {
	var name string
	var su bool
	var b []string
	switch a.Name {
	case "docker logs":
		name = cmdDocker
		su = settings.ForwardPort >= 0
		b, err = a.DockerLogs()
	case "clean ci":
		name = "find"
		b, err = a.CleanCI()
	default:
		err = errors.New("name unsupported")
	}
	if err == nil {
		cmd = exec.CommandContext(ctx, name, b...)
		cmd.Env = a.Env
		if su {
			setUser(cmd)
		}
		setEnv(cmd)
	}
	return
}

func Command(w http.ResponseWriter, r *http.Request) {
	var a Action
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}

	// proxy
	if settings.ForwardPort > 0 && a.Name == "docker logs" {
		if a.Args == nil {
			colonyutil.WriteErr(w, "400", "args required")
			return
		}
		var v ActionDockerLogs
		err := json.Unmarshal(*a.Args, &v)
		if err != nil {
			colonyutil.WriteErr(w, "400", err.Error())
			return
		}
		if v.Host == "" {
			colonyutil.WriteErr(w, "400", "host required")
			return
		}
		b, err := json.Marshal(a)
		if err != nil {
			colonyutil.WriteErr(w, "500", err.Error())
			return
		}
		u := fmt.Sprintf("http://%s:%d/api/command", v.Host, settings.ForwardPort)
		res, err := http.DefaultClient.Post(u, "application/json; charset=utf-8", bytes.NewReader(b))
		if err != nil {
			colonyutil.WriteErr(w, "500", err.Error())
			return
		}
		defer res.Body.Close()
		w.WriteHeader(res.StatusCode)
		_, err = io.Copy(w, res.Body)
		if err != nil {
			colonyutil.WriteErr(w, "500", err.Error())
			return
		}
		return
	}

	ch := make(chan struct{})
	defer close(ch)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ch:
			return
		case <-time.After(5 * time.Minute):
			logrus.Errorln("timeout")
			cancel()
		}
	}()

	cmd, err := a.CommandContext(ctx)
	if err != nil {
		colonyutil.WriteErr(w, "500", err.Error())
		return
	}
	now := time.Now()
	fmt.Fprint(w, now.Format(time.RFC3339Nano)+"\n\n")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		fmt.Fprint(w, "\n"+err.Error()+"\n")
	} else {
		fmt.Fprint(w, "\nexit status 0\n")
	}
	fmt.Fprint(w, time.Since(now).String()+"\n")
}
