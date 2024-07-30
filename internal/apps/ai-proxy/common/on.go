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

package common

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// On .
type On struct {
	Key      string `json:"key" yaml:"key"`
	Operator string `json:"operator" yaml:"operator"`
	Value    string `json:"value" yaml:"value"`
}

func (on *On) On(header http.Header) (bool, error) {
	switch {
	case strings.EqualFold(on.Operator, "exist"):
		return header.Get(on.Key) != "", nil
	case strings.EqualFold(on.Operator, "="):
		return header.Get(on.Key) == on.Value, nil
	default:
		return false, errors.Errorf("invalid operator: %s", on.Operator)
	}
}
