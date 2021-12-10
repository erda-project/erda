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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var CONFIG = command.Command{
	Name:      "config",
	ShortHelp: "show config file for Erda CLI",
	Example:   "$ erda-cli config",
	Run:       ConfigRead,
}

func ConfigRead(ctx *command.Context) error {
	return configOps("inspect", true)
}

func configOps(ops string, noHeaders bool) error {
	_, conf, err := command.GetConfig()
	if err == dicedir.NotExist {
		return errors.New("Please use 'erda-cli config-set' command to set configurations first")
	}
	if err != nil {
		return err
	}

	switch ops {
	case "inspect":
		return configInspect(conf)
	case "current-context":
		return getCurrentContext(conf)
	case "get-contexts":
		return getContexts(conf, noHeaders)
	case "get-platforms":
		return getPlatforms(conf, noHeaders)
	default:
		return errors.New(ops + " ops not found")
	}

	return nil
}

func getPlatforms(conf *command.Config, noHeaders bool) error {
	var data [][]string
	for _, p := range conf.Platforms {
		orgInfo := ""
		if p.OrgInfo != nil {
			orgInfo = fmt.Sprintf("%d/%s/%s", p.OrgInfo.ID, p.OrgInfo.Name, p.OrgInfo.Desc)
		}
		data = append(data, []string{
			p.Name,
			p.Server,
			orgInfo,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"name", "server", "orginfo(id/name/desc)",
		})
	}
	return t.Data(data).Flush()
}

func getContexts(conf *command.Config, noHeaders bool) error {
	var data [][]string
	for _, c := range conf.Contexts {
		var current string
		if c.Name == conf.CurrentContext {
			current = "*"
		}

		data = append(data, []string{
			current,
			c.Name,
			c.PlatformName,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"current", "name", "platform",
		})
	}
	return t.Data(data).Flush()
}

func getCurrentContext(conf *command.Config) error {
	if conf.CurrentContext != "" {
		fmt.Println(conf.CurrentContext)
	} else {
		fmt.Println("not set current context!")
	}

	return nil
}

func configInspect(conf *command.Config) error {
	if conf.Version != command.Version {
		return errors.New(" Version mismatch, should be " + command.Version)
	}

	c, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	fmt.Println(string(c))

	return nil
}
