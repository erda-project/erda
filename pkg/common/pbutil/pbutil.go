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

package pbutil

import (
	"math"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetBool(v *bool) (vv bool, set bool) {
	if v == nil {
		return false, false
	}
	return *v, true
}
func MustGetBool(v *bool) bool {
	if vv, set := GetBool(v); set {
		return vv
	}
	return false
}

func GetUint64(v *uint64) (vv uint64, set bool) {
	if v == nil {
		return 0, false
	}
	return *v, true
}
func MustGetUint64(v *uint64) uint64 {
	if vv, set := GetUint64(v); set {
		return vv
	}
	return 0
}

func GetTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func GetTimeInLocal(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime().In(time.Local)
	return &v
}

func TimeFromMillis(ms uint64) (time.Time, bool) {
	if ms == 0 {
		return time.Time{}, false
	}
	if ms > math.MaxInt64 {
		return time.Time{}, false
	}
	t := time.UnixMilli(int64(ms)).UTC()
	if err := timestamppb.New(t).CheckValid(); err != nil {
		return time.Time{}, false
	}
	return t, true
}
