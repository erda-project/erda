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

package utils

import (
	"fmt"
	"math"
	stdtime "time"

	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/time"
)

func GetInterval(startTimeMs, endTimeMs int64, minInterval stdtime.Duration, preferredPoints int64) string {
	interval := (endTimeMs - startTimeMs) / preferredPoints
	v, unit := time.AutomaticConversionUnit(math.Max(float64(interval*1e6), float64(minInterval.Nanoseconds())))
	return fmt.Sprintf("%s%s", strutil.String(v), unit)
}
