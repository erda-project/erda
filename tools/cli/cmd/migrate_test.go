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

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterMySQLTLS(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		assert.NoError(t, registerMySQLTLS("", ""))
	})

	t.Run("missing CA path", func(t *testing.T) {
		assert.Error(t, registerMySQLTLS("mysql", ""))
	})

	t.Run("invalid CA", func(t *testing.T) {
		caPath := filepath.Join(t.TempDir(), "ca.pem")
		assert.NoError(t, os.WriteFile(caPath, []byte("not a PEM certificate"), 0600))
		assert.Error(t, registerMySQLTLS("rds-invalid-cli", caPath))
	})
}
