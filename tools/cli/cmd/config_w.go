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
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/pkg/errors"
)

var CONFIGW = command.Command{
	Name:      "config-set",
	ShortHelp: "operate config for Erda CLI",
	Example:   `erda config`,
	Args: []command.Arg{
		command.StringArg{}.Name("write-ops"),
		command.StringArg{}.Name("name"),
	},
	Flags: []command.Flag{
		command.StringFlag{Short: "s", Name: "server", Doc: "the http endpoint for openapi of platform", DefaultValue: "https://openapi.erda.cloud"},
		command.StringFlag{Short: "o", Name: "org", Doc: "a org under the platform", DefaultValue: ""},
		command.StringFlag{Short: "e", Name: "platform", Doc: "the name of platform", DefaultValue: ""},

	},
	Run: ConfigOpsW,
}

func ConfigOpsW(ctx *command.Context, ops, name, server, org, platform string) error {
	file, conf, err := command.GetConfig()
	if err != nil {
		return err
	}
	switch ops {
	case "set-platform":
		setPlatform(conf, name, server, org)
	case "set-context":
		setContext(conf, name, platform)
	case "use-context":
		err = useContext(conf, name)
		if err != nil {
			return err
		}
	case "delete-platform":
		deletePlatform(conf, name)
	case "delete-context":
		deleteContext(conf, name)
	default:
		return errors.New(ops + " ops not found")
	}

	err = command.SetConfig(file, conf)
	if err != nil {
		return err
	}

	return nil
}

func setPlatform(conf *command.Config, name, server, org string) {
	notExist := true
	for _, p := range conf.Platforms {
		if p.Name == name {
			p.Server = server
			notExist = false
		}
	}

	if notExist {
		conf.Platforms = append(conf.Platforms, &command.Platform{
			name, server,
			&command.OrgInfo{Name: org},
		})
	}
}

func deletePlatform(conf *command.Config, name string) {
	var ps []*command.Platform
	for _, p := range conf.Platforms {
		if p.Name != name {
			ps = append(ps, p)
		}
	}
	conf.Platforms = ps
}

func setContext(conf *command.Config, name, platform string) {
	notExist := true
	for _, c := range conf.Contexts {
		if c.Name == name {
			c.PlatformName = platform
			notExist = false
		}
	}

	if notExist {
		conf.Contexts = append(conf.Contexts, &command.Ctx {name, platform })
	}
}

func deleteContext(conf *command.Config, name string) {
	var cs []*command.Ctx
	for _, c := range conf.Contexts {
		if c.Name != name {
			cs = append(cs, c)
		}
	}
	conf.Contexts = cs
}

func useContext(conf *command.Config, name string) error {
	for _, c := range conf.Contexts {
		if c.Name == name {
			conf.CurrentContext = name
			return nil
		}
	}

	return errors.New(fmt.Sprintf("context %s not found", name))
}
