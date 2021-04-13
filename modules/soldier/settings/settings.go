// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
