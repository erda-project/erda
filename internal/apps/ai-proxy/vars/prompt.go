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

package vars

import (
	"strings"
)

const (
	UserPromptStart = "|start|"
	UserPromptEnd   = "|end|"
)

func UnwrapUserPrompt(v string) string {
	return strings.TrimPrefix(strings.TrimSuffix(v, UserPromptEnd), UserPromptStart)
}

func WrapUserPrompt(v string) string {
	return UserPromptStart + v + UserPromptEnd
}
