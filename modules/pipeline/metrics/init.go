package metrics

// import (
// 	"github.com/sirupsen/logrus"

// 	"terminus.io/dice/telemetry/metrics"
// )

// // disableMetrics 是否禁用 metric 相关操作
// var disableMetrics bool

// var bulkClient *metrics.BulkAction

// func Initialize() {
// 	client := metrics.NewClient()
// 	bulk, err := client.NewBulk()
// 	if err != nil {
// 		logrus.Errorf("[alert] failed to init event bulk client, disable metric report, err: %v", err)
// 		disableMetrics = true
// 		return
// 	}
// 	bulkClient = bulk
// 	logrus.Info("metrics enabled")
// }
