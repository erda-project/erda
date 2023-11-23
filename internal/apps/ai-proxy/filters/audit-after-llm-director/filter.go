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

package audit_after_llm_director

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "audit-after-llm-director"
)

var (
	_ reverseproxy.ActualRequestFilter = (*Filter)(nil)
	_ reverseproxy.ResponseFilter      = (*Filter)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Filter struct {
	*reverseproxy.DefaultResponseFilter

	firstResponseAt time.Time
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Filter{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}
