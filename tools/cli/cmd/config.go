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
)

var CONFIG = command.Command{
	Name:      "config",
	ShortHelp: "operate config for Erda CLI",
	Example:   `erda config`,
	Args: []command.Arg{
		command.StringArg{}.Name("ops"),
	},
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers",
			Doc:          "When using the default or custom-column output format, don't print headers (default print headers)",
			DefaultValue: false},
	},
	Run: ConfigOps,
}

func ConfigOps(ctx *command.Context, ops string, noHeaders bool) error {
	switch ops {
	case "inspect":
		return configInspect(ctx)
	case "current-context":
		return getCurrentContext(ctx)
	case "get-contexts":
		return getContexts(ctx, noHeaders)
	case "get-platforms":
		return getPlatforms(ctx, noHeaders)
	default:
		return errors.New(ops + " ops not found")
	}

	return nil
}

func getPlatforms(ctx *command.Context, noHeaders bool) error {
	_, conf, err := command.GetConfig()
	if err != nil {
		return err
	}

	var data [][]string
	for _, p := range conf.Platforms {
		data = append(data, []string{
			p.Name,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"name",
		})
	}
	return t.Data(data).Flush()
}

func getContexts(ctx *command.Context, noHeaders bool) error {
	_, conf, err := command.GetConfig()
	if err != nil {
		return err
	}

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

func getCurrentContext(ctx *command.Context) error {
	_, conf, err := command.GetConfig()
	if err != nil {
		return err
	}

	if conf.CurrentContext != "" {
		fmt.Println(conf.CurrentContext)
	} else {
		fmt.Println("not set current context!")
	}

	return nil
}

func configInspect(ctx *command.Context) error {
	_, conf, err := command.GetConfig()
	if err != nil {
		return err
	}

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
