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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/sysconf"
)

func StartCron() {
	logrus.Info("start cron")
	t := time.NewTicker(time.Minute)
	for {
		if err := Cron(false); err != nil {
			logrus.Errorf("cron failed: %s", err)
		}
		<-t.C
	}
}

// Cron Download automatic operation and maintenance scripts and start cron tasks
func Cron(reload bool) error {
	m := make(map[string]*Action)
	err := filepath.Walk(scriptsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == scriptsPath {
			return nil
		}
		if info.IsDir() && filepath.Dir(path) == scriptsPath {
			a := NewAction(path)
			if err := a.ReadEnv(); err != nil {
				return fmt.Errorf("read %s env failed: %s", a.Name, err.Error())
			}
			m[a.Name] = a
		}
		return filepath.SkipDir
	})
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, a := range actions {
		if a.Cron != nil {
			a.Cron.Stop()
		}
		if a.CancelFunc != nil {
			a.CancelFunc()
		}
	}
	actions = m
	for _, a := range actions {
		func(a *Action) {
			a.Path = scriptsPath
			logrus.Infoln(a.Name, a.CronTime, a.Nodes)
			if a.CronTime != "" {
				a.Cron = cron.New()
				err := a.Cron.AddFunc(a.CronTime, func() {
					if err := a.Run(NetdataEnv()); err != nil {
						logrus.Errorf("run %s failed: %s", a.Name, err.Error())
						return
					}
				})
				if err != nil {
					logrus.Errorf("%s cron failed: %s", a.Name, err.Error())
				}
				a.Cron.Start()
			}
		}(a)
	}
	return nil
}

var (
	mutex       sync.Mutex
	actions     = make(map[string]*Action)
	spaceRegexp = regexp.MustCompile(`\s+`)
)

func NetdataEnv() map[string]string {
	m := make(map[string]string)
	m["NETDATA_CI_PATH"] = os.Getenv("NETDATA_CI_PATH")
	m["NETDATA_AGENT_PATH"] = os.Getenv("NETDATA_AGENT_PATH")
	return m
}

// Run Running tools
func (a *Action) Run(m map[string]string) error {
	mutex.Lock()
	if a.Running {
		mutex.Unlock()
		return fmt.Errorf("running")
	}
	a.Running, a.RunTime = true, time.Now()
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		a.Running = false
		mutex.Unlock()
	}()

	logrus.Infof("start %s %s", a.Name, a.RunTime.Format(time.RFC3339Nano))
	var err error
	//a.ClusterInfo, err = readCluster()
	//if err != nil {
	//	logrus.Errorf("read cluster failed: %s", err.Error())
	//	return err
	//}
	//if scriptBlacklist, ok := a.ClusterInfo.Settings["scriptBlacklist"].(string); ok && scriptBlacklist != "" {
	//	for _, s := range strings.Split(scriptBlacklist, ",") {
	//		if s = strings.TrimSpace(s); s == "ALL" || s == a.Name {
	//			return errors.New("blacklist")
	//		}
	//	}
	//}

	if err = a.ReadEnv(); err != nil {
		logrus.Errorf("read env failed: %s", err.Error())
		return err
	}

	a.Context, a.CancelFunc = context.WithTimeout(context.Background(), time.Hour)
	defer func() {
		a.CancelFunc()
		a.Context, a.CancelFunc = nil, nil
	}()

	var b bytes.Buffer
	for k, v := range a.Env {
		if s, ok := m[k]; ok {
			v = quoting(s)
		}
		b.WriteString(k + "=" + v + "\n")
	}
	err = ioutil.WriteFile(a.File("env.sh"), b.Bytes(), 0644)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(logsPath, a.Name+".log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	//jo := &jsonOutput{w: f}

	//if e := a.UpdateTaskInfo(true, false, ""); e != nil {
	//	logrus.Warning(e)
	//}

	path := filepath.Join(a.Path, a.Name)

	var content bytes.Buffer

	fmt.Fprintf(f, "start %s %s\n\n", a.Name, a.RunTime.Format(time.RFC3339Nano))
	//if epush := pushToCollector(a.ClusterInfo.Name,
	//	fmt.Sprintf("start %s %s\n\n", a.Name, a.RunTime.Format(time.RFC3339Nano)),
	//	"stdout", "info"); epush != nil {
	//	logrus.Errorln(epush)
	//}
	err = func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("%+v", e)
			}
		}()
		defer func() {
			//if epush := pushToCollector(a.ClusterInfo.Name, content.String(), "stdout", "info"); epush != nil {
			//	logrus.Errorln(epush)
			//}
			_, err := f.WriteString(content.String())
			if err != nil {
				logrus.Errorln("failed to write script log to file")
			}
		}()
		if a.Nodes != "" {
			var hosts []string
			switch a.Nodes {
			case "@all":
				for _, node := range a.ClusterInfo.System.Nodes {
					switch node.Type {
					case sysconf.NodeTypeMaster,
						sysconf.NodeTypePublic, sysconf.NodeTypeLB,
						sysconf.NodeTypePrivate, sysconf.NodeTypeApp:
						hosts = append(hosts, node.IP)
					}
				}
			case "@one":
				for _, node := range a.ClusterInfo.System.Nodes {
					if node.Type == "private" || node.Type == "app" {
						hosts = append(hosts, node.IP)
						break
					}
				}
			case "@master",
				"@public", "@lb",
				"@private", "@app":
				nt := a.Nodes[1:]
				for _, node := range a.ClusterInfo.System.Nodes {
					if node.Type == nt {
						hosts = append(hosts, node.IP)
					}
				}
			case "@slave":
				for _, node := range a.ClusterInfo.System.Nodes {
					if node.Type == sysconf.NodeTypePublic || node.Type == sysconf.NodeTypePrivate ||
						node.Type == sysconf.NodeTypeLB || node.Type == sysconf.NodeTypeApp {
						hosts = append(hosts, node.IP)
					}
				}
			default:
				hosts = spaceRegexp.Split(strings.TrimSpace(a.Nodes), -1)
			}
			for i, s := range hosts {
				ip := net.ParseIP(s)
				if ip == nil {
					return fmt.Errorf("invalid node: %s", s)
				}
				s = ip.String()
				hosts[i] = fmt.Sprintf("%s:%d", s, a.ClusterInfo.System.SSH.Port)
				for j := i - 1; j >= 0; j-- {
					if hosts[i] == hosts[j] {
						return fmt.Errorf("duplicate node: %s", s)
					}
				}
			}

			var args, env []string
			if a.ClusterInfo.System.SSH.PrivateKey != "" && a.ClusterInfo.System.SSH.Account != "" {
				args = []string{"-u", a.ClusterInfo.System.SSH.Account, "-k", "."}
				env = []string{"ORGALORG_KEY=" + a.ClusterInfo.System.SSH.PrivateKey}
			} else if a.ClusterInfo.System.SSH.Password != "" && a.ClusterInfo.System.SSH.User != "" {
				args = []string{"-u", a.ClusterInfo.System.SSH.User, "-p"}
				env = []string{"ORGALORG_PASSWORD=" + a.ClusterInfo.System.SSH.Password}
			} else {
				return fmt.Errorf("no ssh auth method")
			}
			args = append(args, "-s", "-x", "-y", "--color", "never", "--json")
			idx := len(args)

			run := func(hosts []string) error {
				s := strings.Join(hosts, "\n")
				d := filepath.Join("/tmp", "autoop_"+a.Name)

				args = args[:idx]
				env = env[:1]
				{ // prepare
					args = append(args, "-C", "rm", "-rf", d)
					c := exec.CommandContext(a.Context, "orgalorg", args...)
					c.Dir = path
					c.Stdin = strings.NewReader(s)
					c.Stdout = &content
					c.Stderr = &content
					c.Env = env
					setEnv(c)
					if e := c.Run(); e != nil {
						//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("prepare: %s", e.Error()),
						//	"stderr", "error"); epush != nil {
						//	logrus.Errorln(epush)
						//}
						return fmt.Errorf("prepare: %s", e.Error())
					}
				}

				args = args[:idx]
				env = env[:1]
				{ // upload
					args = append(args, "-e", "-r", d, "-U", ".")
					c := exec.CommandContext(a.Context, "orgalorg", args...)
					c.Dir = path
					c.Stdin = strings.NewReader(s)
					c.Stdout = &content
					c.Stderr = &content
					c.Env = env
					setEnv(c)
					if e := c.Run(); e != nil {
						//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("upload: %s", e.Error()),
						//	"stderr", "error"); epush != nil {
						//	logrus.Errorln(epush)
						//}
						return fmt.Errorf("upload: %s", e.Error())
					}
				}

				args = args[:idx]
				env = env[:1]
				{ // run
					args = append(args, "-C", "cd", d, "&&", "bash", "run.sh")
					c := exec.CommandContext(a.Context, "orgalorg", args...)
					c.Dir = path
					c.Stdin = strings.NewReader(s)
					c.Stdout = &content
					c.Stderr = &content
					c.Env = env
					setEnv(c)
					if e := c.Run(); e != nil {
						//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("run: %s", e.Error()),
						//	"stderr", "error"); epush != nil {
						//	logrus.Errorln(epush)
						//}
						return fmt.Errorf("run: %s", e.Error())
					}
				}

				return nil
			}
			switch len(hosts) {
			case 0:
				return fmt.Errorf("no nodes")
			case 1:
				return run(hosts)
			default:
				if e := run(hosts[:1]); e != nil {
					return e
				}
				return run(hosts[1:])
			}
		} else {
			c := exec.CommandContext(a.Context, "bash", "run.sh")
			c.Dir = path
			c.Stdout = &content
			c.Stderr = &content
			setEnv(c)
			if e := c.Run(); e != nil {
				//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("run: %s", e.Error()),
				//	"stderr", "error"); epush != nil {
				//	logrus.Errorln(epush)
				//}
				return fmt.Errorf("run: %s", e.Error())
			}
		}
		return nil
	}()
	if d := time.Since(a.RunTime); err != nil {
		logrus.Errorf("end %s %s failed: %s", a.Name, d, err.Error())
		//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("end %s %s failed: %s",
		//	a.Name, d, err.Error()), "stderr", "error"); epush != nil {
		//	logrus.Errorln(epush)
		//}
		fmt.Fprintf(f, "\n\nend %s %s failed: %s\n", a.Name, d, err.Error())

		//jo.a = append(jo.a, apistructs.AutoopOutputLine{Stream: "stderr", Body: err.Error()})
		//b, e := json.Marshal(jo.a)
		//if e != nil {
		//	logrus.Warning(e)
		//	b = []byte("[]")
		//}
		//e = a.UpdateTaskInfo(false, false, string(b))
		//if e != nil {
		//	logrus.Warning(e)
		//}
	} else {
		logrus.Infof("end %s %s succeed", a.Name, d)
		//if epush := pushToCollector(a.ClusterInfo.Name, fmt.Sprintf("end %s %s succeed", a.Name, d),
		//	"stdout", "info"); epush != nil {
		//	logrus.Errorln(epush)
		//}
		fmt.Fprintf(f, "\n\nend %s %s succeed\n", a.Name, d)
		//
		//if e := a.UpdateTaskInfo(false, true, ""); e != nil {
		//	logrus.Warning(e)
		//}
	}
	return err
}

func setEnv(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(cmd.Env, "TERM=xterm-256color", "LANG=en_US.UTF-8")
}

//func pushToCollector(cluster string, content string, stream string, level string) error {
//	log := NewLogContent()
//	log.Stream = stream
//	log.ID = cluster + "-script"
//	log.Tags = Tag{
//		Level: level,
//	}
//	log.Content = content
//	log.TimeStamp = time.Now().UnixNano()
//	response, err := httpclient.New().Post(conf.CollectorURL()).
//		Path("/collect/logs/job").JSONBody(&[]LogContent{log}).Do().DiscardBody()
//	if err != nil {
//		return errors.New(fmt.Sprintf("failed to push script log to collect, reason: %s", err.Error()))
//	}
//	if !response.IsOK() {
//		return errors.New(fmt.Sprintf("failed to push script log to collect, status code %d",
//			response.StatusCode()))
//	}
//	return nil
//}
