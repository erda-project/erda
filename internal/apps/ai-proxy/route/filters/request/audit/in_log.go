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

package audit

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/audit/audit_util"
)

func (f *Audit) requestInLog(in *http.Request) error {
	// Decide whether to dump body based on content-type
	contentType := in.Header.Get("Content-Type")
	shouldDumpBody := audit_util.ShouldDumpBody(contentType)

	dumpBytesIn, err := httputil.DumpRequest(in, shouldDumpBody)
	if err != nil {
		return fmt.Errorf("failed to dump request in, err: %v", err)
	}

	logger := ctxhelper.MustGetLogger(in.Context())
	if shouldDumpBody {
		logger.Infof("audit proxy request in:\n%s", dumpBytesIn)
	} else {
		logger.Infof("audit proxy request in (body omitted due to binary content-type: %s):\n%s",
			contentType, dumpBytesIn)
	}

	return nil
}
