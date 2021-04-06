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
