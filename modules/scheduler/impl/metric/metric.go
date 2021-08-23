// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
