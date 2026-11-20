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

package mysqltls

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda-infra/pkg/mysqldriver"
)

// Open registers a named MySQL TLS profile. Driver built-in modes keep their
// existing behavior and do not accept certificate paths.
func Open(tlsName, caCertPath, clientCertPath, clientKeyPath string) error {
	if tlsName == "" {
		if caCertPath != "" || clientCertPath != "" || clientKeyPath != "" {
			return fmt.Errorf("MYSQL_TLS is required when MySQL certificate paths are configured")
		}
		return nil
	}

	if isBuiltIn(tlsName) {
		if caCertPath != "" || clientCertPath != "" || clientKeyPath != "" {
			return fmt.Errorf("MySQL certificate paths require a custom MYSQL_TLS name")
		}
		return nil
	}

	if caCertPath == "" {
		return fmt.Errorf("MYSQL_CACERTPATH is required when custom MYSQL_TLS is configured")
	}
	return mysqldriver.OpenTLS(tlsName, caCertPath, clientCertPath, clientKeyPath)
}

func isBuiltIn(tlsName string) bool {
	switch strings.ToLower(tlsName) {
	case "true", "false", "skip-verify", "preferred":
		return true
	default:
		return false
	}
}
