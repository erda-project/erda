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

package token

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/erda-project/erda/internal/tools/gittar/pkg/util/guid"
)

var (
	adminAuthToken string
	once           sync.Once
)

// InitAdminAuthToken initializes the internal api token.
// If the token file exists, read it.
// If not, generate a new one and write it to the file.
func InitAdminAuthToken(path string) (string, error) {
	var err error
	once.Do(func() {
		if path == "" {
			err = errors.New("auth token path is empty")
			return
		}

		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return
		}

		// try read
		if _, statErr := os.Stat(path); statErr == nil {
			content, readErr := ioutil.ReadFile(path)
			if readErr != nil {
				err = readErr
				return
			}
			adminAuthToken = strings.TrimSpace(string(content))
			if adminAuthToken != "" {
				return
			}
		}

		// generate new
		adminAuthToken = guid.NewString()
		err = ioutil.WriteFile(path, []byte(adminAuthToken), 0600)
	})
	return adminAuthToken, err
}

// GetAdminAuthToken returns the current internal api token.
func GetAdminAuthToken() string {
	return adminAuthToken
}

// ValidateAdminAuthToken checks if the bearer token matches the internal token.
// bearerToken should be strictly the token string (without "Bearer " prefix if handled outside,
// but usually it's easier to pass the clean token here).
func ValidateAdminAuthToken(auth string) bool {
	return adminAuthToken != "" && auth == adminAuthToken
}
