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

package pbconvert

import (
	"fmt"

	st "github.com/golang/protobuf/ptypes/struct"
)

// TODO. slice, map, struct support?
func ToValue(v interface{}) *st.Value {
	switch v := v.(type) {
	case nil:
		return nil
	case bool:
		return &st.Value{
			Kind: &st.Value_BoolValue{
				BoolValue: v,
			},
		}
	case int:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int8:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint8:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: v,
			},
		}
	case string:
		return &st.Value{
			Kind: &st.Value_StringValue{
				StringValue: v,
			},
		}
	default:
		return &st.Value{
			Kind: &st.Value_StringValue{
				StringValue: fmt.Sprintf("%+v", v),
			},
		}
	}
}
