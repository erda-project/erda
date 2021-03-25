package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_GetLog(*testing.T) {
//	os.Setenv("MONITOR_ADDR", "monitor.default.svc.cluster.local:7096")
//	b := New(WithMonitor())
//
//	fmt.Println(b.GetLog(apistructs.DashboardSpotLogRequest{
//		ID:     "pipeline-task-244",
//		Source: apistructs.DashboardSpotLogSourceJob,
//		Stream: apistructs.DashboardSpotLogStreamStdout,
//		Count:  -50,
//		Start:  0,
//		End:    time.Duration(1590047806647571944),
//	}))
//}
