package actionagent

import (
	"fmt"
	"testing"
)

//func TestWatchFiles(t *testing.T) {
//	logrus.SetLevel(logrus.DebugLevel)
//	logrus.Debug("begin watch files")
//
//	fw, err := filewatch.New()
//	assert.NoError(t, err)
//	ctx, cancel := context.WithCancel(context.Background())
//	agent := Agent{
//		EasyUse: EasyUse{
//			EnablePushLog2Collector: true,
//			CollectorAddr:           "collector.default.svc.cluster.local:7076",
//			TaskLogID:               "agent-push-1",
//
//			RunMultiStdoutFilePath: "/tmp/stdout",
//			RunMultiStderrFilePath: "/tmp/stderr",
//		},
//		Ctx:         ctx,
//		Cancel:      cancel,
//		FileWatcher: fw,
//	}
//
//	go agent.watchFiles()
//
//	// mock logic done
//	time.Sleep(time.Millisecond * 10) // very fast action done
//
//	// teardown
//	agent.PreStop()
//
//	time.Sleep(time.Second * 1)
//}

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
