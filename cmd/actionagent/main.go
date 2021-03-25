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
		Ctx:               ctx,
		Cancel:            cancel,
	}
	agent.Execute(bytes.NewBufferString(args))
	agent.Teardown()
	agent.Exit()
}
