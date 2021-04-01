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
