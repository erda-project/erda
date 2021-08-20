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

package main

import (
	"bytes"
	"context"
	"os"

	"github.com/sirupsen/logrus"

	_ "github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/modules/actionagent"
)

type PlatformLogFormmater struct {
	logrus.TextFormatter
}

func (f *PlatformLogFormmater) Format(entry *logrus.Entry) ([]byte, error) {
	_bytes, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	return append([]byte("[Platform Log] "), _bytes...), nil
}

func main() {
	args := os.Args[1]
	realMain(args)
}

func realMain(args string) {
	// set logrus
	logrus.SetFormatter(&PlatformLogFormmater{
		logrus.TextFormatter{
			ForceColors:            true,
			DisableTimestamp:       true,
			DisableLevelTruncation: false,
			PadLevelText:           false,
		},
	})
	logrus.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		logrus.Fatal("failed to run action: no args passed in")
	}
	ctx, cancel := context.WithCancel(context.Background())
	agent := &actionagent.Agent{
		Errs:              make([]error, 0),
		PushedMetaFileMap: make(map[string]string),
		TextBlackList:         make([]string, 0), // enciphered data will Replaced by '******' when log output
		Ctx:               ctx,
		Cancel:            cancel,
	}
	agent.Execute(bytes.NewBufferString(args))
	agent.Teardown()
	agent.Exit()
}
