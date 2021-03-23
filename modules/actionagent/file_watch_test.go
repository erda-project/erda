package actionagent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/actionagent/filewatch"
)

func TestWatchFiles(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("begin watch files")

	fw, err := filewatch.New()
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	agent := Agent{
		EasyUse: EasyUse{
			EnablePushLog2Collector: true,
			CollectorAddr:           "collector.default.svc.cluster.local:7076",
			TaskLogID:               "agent-push-1",

			RunMultiStdoutFilePath: "/tmp/stdout",
			RunMultiStderrFilePath: "/tmp/stderr",
		},
		Ctx:         ctx,
		Cancel:      cancel,
		FileWatcher: fw,
	}

	go agent.watchFiles()

	// mock logic done
	time.Sleep(time.Second * 1) // very fast action done

	// teardown
	agent.Teardown(0)

	time.Sleep(time.Second * 15)
}

func TestSlice(t *testing.T) {
	f := func(flogs *[]string, line string) {
		*flogs = append(*flogs, line)
	}
	var logs = &[]string{}
	f(logs, "l1")
	fmt.Println(logs)
	f(logs, "l2")
	fmt.Println(logs)
}
