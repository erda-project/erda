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
