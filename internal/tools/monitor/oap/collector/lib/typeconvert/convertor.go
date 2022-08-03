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

package typeconvert

import (
	"fmt"
	"strconv"
)

func ToInt64(obj interface{}) (int64, error) {
	switch val := obj.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	return 0, fmt.Errorf("invalid type to int64")
}

func ToFloat64(obj interface{}) (float64, error) {
	switch val := obj.(type) {
	case int:
		return float64(int64(val)), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	case string:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	return 0, fmt.Errorf("invalid type to float64")
}
