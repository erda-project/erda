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

package audithelper

import (
	"os"
	"strconv"
	"strings"
)

const envKeyAuditWriteDisabled = "AI_PROXY_AUDIT_WRITE_DISABLED"

func AuditWriteDisabled() bool {
	raw, ok := os.LookupEnv(envKeyAuditWriteDisabled)
	if !ok {
		return false
	}
	disabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	return err == nil && disabled
}
