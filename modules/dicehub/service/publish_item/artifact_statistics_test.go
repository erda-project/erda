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

package publish_item

import (
	"fmt"
	"sort"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSlice(t *testing.T) {
	//result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", (float64(60267-78686)/float64(78686))*100.0), 64)

	recentAvgActiveTime := 60267
	lastAvgActiveTime := 78686
	fmt.Println((float64(60267-78686) / float64(78686)) * 100.0)
	fmt.Println((float64(recentAvgActiveTime-lastAvgActiveTime) / float64(lastAvgActiveTime)) * 100.0)

	logrus.Infof("monitor request params 往前推7天 recentAvgActiveTime: %d,lastAvgActiveTime: %d", recentAvgActiveTime, lastAvgActiveTime)
	logrus.Infof("monitor request params 往前推7天 result: %f", (float64(recentAvgActiveTime-lastAvgActiveTime)/float64(lastAvgActiveTime))*100.0)
	logrus.Infof("monitor request params 往前推7天 resultStr: %s", fmt.Sprintf("%.2f", (float64(recentAvgActiveTime-lastAvgActiveTime)/float64(lastAvgActiveTime))*100.0))

}

func TestMetrics(t *testing.T) {
	strslice := []string{"1.2.3", "1.1.2", "1.1.9", "1.3.1"}

	sort.Sort(sort.StringSlice(strslice))

	fmt.Println(strslice)
}
