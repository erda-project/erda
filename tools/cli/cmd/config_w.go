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

	"github.com/erda-project/erda/tools/cli/dicedir"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

var CONFIGW = command.Command{
	Name:      "config-set",
	ShortHelp: "Write config file for Erda CLI",
	Example:   `$ erda-cli config-set <set-platform|set-context|use-context|delete-platform|delete-context> <name> [flags]`,
	Args: []command.Arg{
		command.StringArg{}.Name("write-ops"),
		command.StringArg{}.Name("name"),
	},
	Flags: []command.Flag{
		command.StringFlag{Short: "", Name: "server", Doc: "The http endpoint for openapi of platform", DefaultValue: "https://openapi.erda.cloud"},
		command.StringFlag{Short: "", Name: "org", Doc: "An org under the platform", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "platform", Doc: "The name of platform", DefaultValue: ""},
	},
	Run: ConfigOpsW,
}

func ConfigOpsW(ctx *command.Context, ops, name, server, org, platform string) error {
	file, conf, err := command.GetConfig()
	if err != nil && err != dicedir.NotExist {
		return err
	}
	switch ops {
	case "set-platform":
		if server == "" {
			return errors.New("Must set server by --server")
		}
		setPlatform(conf, name, server, org)
	case "set-context":
		if platform == "" {
			return errors.New("Must set platform by --platform")
		}
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
		conf.Contexts = append(conf.Contexts, &command.Ctx{name, platform})
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
