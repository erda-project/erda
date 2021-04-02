package metric

//import (
//	"terminus.io/dice/telemetry/promxp"
//)
//
//var (
//	ServiceCreateTotal = "serviceCreateTotal"
//	ServiceCreateError = "serviceCreateError"
//	ServiceRemoveTotal = "serviceRemoveTotal"
//	ServiceRemoveError = "erviceRemoveError"
//	JobCreateTotal     = "jobCreateTotal"
//	JobCreateError     = "jobCreateError"
//	LableTotal         = "lableTotal"
//	LableError         = "lableError"
//)
//var ErrorCounter = promxp.RegisterAutoResetCounterVec(
//	"error_number",
//	"number of error event occur",
//	map[string]string{},
//	[]string{"type"},
//)
//
//var TotalCounter = promxp.RegisterAutoResetCounterVec(
//	"totel_number",
//	"number of event occur",
//	map[string]string{},
//	[]string{"type"},
//)
//
//type Metric struct {
//	ErrorCounter *promxp.AutoResetCounterVec
//	TotalCounter *promxp.AutoResetCounterVec
//}
//
//func New() Metric {
//	return Metric{
//		ErrorCounter: ErrorCounter,
//		TotalCounter: TotalCounter,
//	}
//}
