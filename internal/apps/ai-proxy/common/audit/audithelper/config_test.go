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

import "testing"

func TestSetAuditWriteDisabled(t *testing.T) {
	SetAuditWriteDisabled(false)
	if AuditWriteDisabled() {
		t.Fatalf("expected audit write disabled to be false by default")
	}

	SetAuditWriteDisabled(true)
	if !AuditWriteDisabled() {
		t.Fatalf("expected audit write disabled to be true after set")
	}
}

func TestSetAuditWriteDisabledReset(t *testing.T) {
	SetAuditWriteDisabled(true)
	SetAuditWriteDisabled(false)
	if AuditWriteDisabled() {
		t.Fatalf("expected audit write disabled to be false after reset")
	}
}
