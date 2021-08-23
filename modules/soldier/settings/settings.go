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

package settings

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/erda-project/erda/pkg/discover"
)

var (
	HTTPAddr    = ":9028"
	LogLevel    = logrus.InfoLevel
	PIDFile     = ""
	OpenAPIURL  = ""
	ForwardPort = 0
	DownloadDB  *leveldb.DB
	exitChan                    = make(chan struct{})
	ExitChan    <-chan struct{} = exitChan
)

func LoadEnv() {
	var err error

	if s := os.Getenv("HTTP_ADDR"); s != "" {
		HTTPAddr = s
	}

	if s := os.Getenv("LOG_LEVEL"); s != "" {
		LogLevel, err = logrus.ParseLevel(s)
		if err != nil {
			logrus.Fatalln(err)
		}
	}
	logrus.SetLevel(LogLevel)

	if s := os.Getenv("PID_FILE"); s != "" {
		PIDFile = s
	}
	if PIDFile != "" {
		err = ioutil.WriteFile(PIDFile, []byte(strconv.Itoa(os.Getpid())), 0644)
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	if s := os.Getenv("FORWARD_PORT"); s != "" {
		ForwardPort, err = strconv.Atoi(s)
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	//DownloadDB, err = leveldb.OpenFile("/opt/download.db", nil)
	//if err != nil {
	//	logrus.Fatalln(err)
	//}

	if b, err := strconv.ParseBool(os.Getenv("DICE_IS_EDGE")); err == nil {
		if b {
			OpenAPIURL = os.Getenv("OPENAPI_PUBLIC_URL")
		} else {
			OpenAPIURL = discover.Openapi()
		}
	}
	if OpenAPIURL == "" {
		logrus.Fatalln("OPENAPI_URL")
	}
}

func Wait() {
	nc := make(chan os.Signal, 1)
	signal.Notify(nc, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infoln("wait signal")
	<-nc
	logrus.Infoln("signal notify")

	close(exitChan)
	time.Sleep(time.Second)

	//DownloadDB.Close()

	if PIDFile != "" {
		err := os.Remove(PIDFile)
		if err != nil {
			logrus.Errorln(err)
		}
	}
}
