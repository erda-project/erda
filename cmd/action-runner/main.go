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

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	_ "github.com/erda-project/erda-infra/base/version"
	actionrunner "github.com/erda-project/erda/modules/action-runner"
)

var config = flag.String("config", "./config.json", "file path")

func main() {
	flag.Parse()
	conf := readConfig(*config)
	runner := actionrunner.New(conf)
	err := runner.Run()
	if err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
}

func readConfig(path string) *actionrunner.Conf {
	var conf actionrunner.Conf
	if len(path) > 0 {
		byts, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Error(err)
			os.Exit(-1)
		}
		err = json.Unmarshal(byts, &conf)
		if err != nil {
			logrus.Error(err)
			os.Exit(-1)
		}
	}
	conf.BuildPath = getEnv("BUILD_ROOT_PATH", conf.BuildPath)
	if len(conf.BuildPath) <= 0 {
		conf.BuildPath = os.TempDir()
	}
	conf.OpenAPI = getEnv("OPENAPI_UEL", conf.OpenAPI)
	conf.Token = getEnv("TOKEN", conf.Token)
	if conf.MaxTask < 1 {
		conf.MaxTask = 1
	}
	if conf.FailedTaskKeepHours < 1 {
		conf.FailedTaskKeepHours = 3
	}
	conf.MaxTask = convInt(getEnv("MAX_TASK", strconv.Itoa(conf.MaxTask)))
	return &conf
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if len(val) > 0 {
		return val
	}
	return def
}

func convInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	return val
}
