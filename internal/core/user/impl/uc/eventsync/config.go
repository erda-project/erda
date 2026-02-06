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

package eventsync

type config struct {
	Host         string `file:"host"`
	ClientID     string `file:"client_id"`
	ClientSecret string `file:"client_secret"`

	UCAuditorCron         string `file:"uc_auditor_cron" default:"0 */1 * * * ?"`
	UCAuditorPullSize     uint64 `file:"uc_auditor_pull_size" default:"30"`
	CompensationExecCron  string `file:"compensation_exec_cron" default:"0 */5 * * * ?"`
	UCSyncRecordCleanCron string `file:"uc_syncrecord_clean_cron" default:"0 0 3 * * ?"`

	CompensationBatchSize int `file:"compensation_batch_size" default:"10"`
	CleanRecordDays       int `file:"clean_record_days" default:"7"`
}
