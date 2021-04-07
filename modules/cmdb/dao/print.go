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

package dao

import (
	"time"

	"github.com/sirupsen/logrus"
)

var (
	containerDiffCount = 0
	containerCount     = 0
	hostDiffCount      = 0
	hostCount          = 0
	upTime             = 0
)

// Count 记录cmdb运行时间，获取的容器数量，主机数量
func Count() {
	logrus.Info("start init Container Count....")
	for {
		logrus.Info("Container: ", containerCount)
		logrus.Info("ContainerDiff: ", containerDiffCount)
		logrus.Info("Host: ", hostCount)
		logrus.Info("UpTime: ", upTime)
		upTime++
		time.Sleep(60 * time.Second)
	}
}
