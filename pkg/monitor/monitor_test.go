package monitor

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var _ = func() int {
	os.Setenv("DINGDING", "https://oapi.dingtalk.com/robot/send?access_token=a")
	os.Setenv("DINGDING_pipeline", "https://oapi.dingtalk.com/robot/send?access_token=b")
	return 0
}()

func TestInitMonitor(t *testing.T) {
	logrus.Errorf("[alert] hello [alert]")
	logrus.Errorf("[pipeline] hello [pipeline]")
	time.Sleep(time.Second * 1)
}
