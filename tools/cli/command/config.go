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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/tools/cli/dicedir"
)

var (
	Version = "1.0"
)

type Config struct {
	Version        string      `yaml:"version"`
	Platforms      []*Platform `yaml:"platforms"`
	Contexts       []*Ctx      `yaml:"contexts"`
	CurrentContext string      `yaml:"current_context"`
}

type Platform struct {
	Name    string   `yaml:"name"`
	Server  string   `yaml:"server"`
	OrgInfo *OrgInfo `yaml:"org_info"`
}

type Ctx struct {
	Name         string `yaml:"name"`
	PlatformName string `yaml:"platform_name"`
}

type CurCtx struct {
	Platform Platform
	Name     string
}

func GetCurContext() (CurCtx, error) {
	var cur CurCtx
	_, conf, err := GetConfig()
	if err != nil {
		return cur, err
	}

	if conf.CurrentContext == "" {
		return cur, errors.New("current context not set")
	}

	for _, c := range conf.Contexts {
		if c.Name == conf.CurrentContext {
			cur.Name = c.Name

			if c.PlatformName != "" {
				for _, p := range conf.Platforms {
					if p.Name == c.PlatformName {
						cur.Platform = *p

						return cur, nil
					}
				}
			}
		}
	}

	return cur, errors.New("current context not found")
}

func GetConfig() (string, *Config, error) {
	config, err := dicedir.FindGlobalConfig()
	if err != nil {
		return config, nil, err
	}

	f, err := os.Open(config)
	if err != nil {
		return config, nil, err
	}
	var conf Config
	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		os.Remove(config)
		return config, nil, err
	}

	return config, &conf, nil
}

func SetConfig(file string, conf *Config) error {
	c, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, c, 655)
	if err != nil {
		return err
	}

	return nil
}
