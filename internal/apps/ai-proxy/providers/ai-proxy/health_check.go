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

package ai_proxy

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
)

func (p *provider) HealthCheckAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := p.checkTables(r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success": false, "error": "%v"}`, err)))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success": true}`)))
	}
}

func (p *provider) checkTables(r *http.Request) error {
	tables := []string{
		(&audit.Audit{}).TableName(),
		(&mcp_server.MCPServer{}).TableName(),
	}
	for _, table := range tables {
		var dummy int
		err := p.Dao.Q().
			WithContext(r.Context()).
			Raw(fmt.Sprintf(`SELECT 1 FROM %s LIMIT 1`, table)).
			Scan(&dummy).Error
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
	}

	return nil
}
