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

	"github.com/erda-project/erda/pkg/strutil"
)

func ConcatBearer(v string) string {
	return "Bearer " + v
}

func TrimBearer(v string) string {
	return strutil.TrimPrefixes(v, "Bearer ", "bearer ")
}

func ConcatChunkDataPrefix(v []byte) []byte {
	return []byte("data: " + string(v) + "\n\n")
}

func TrimChunkDataPrefix(v []byte) []byte {
	return []byte(strings.TrimPrefix(strings.TrimSuffix(string(v), "\n\n"), "data: "))
}
