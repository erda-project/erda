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

package migrator

import (
	"fmt"
	"sort"

	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/types"
)

type Equal struct {
	equal  bool
	reason string
}

func (e *Equal) Equal() bool {
	return e.equal
}

func (e *Equal) Reason() string {
	return e.reason
}

func FieldTypeEqual(l, r *types.FieldType) *Equal {
	if l.Tp != r.Tp {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("字段类型不一致(FieldType.Tp), 期望值: %s, 实际值: %s; ", l.String(), r.String()),
		}
	}

	if !(l.Flen == -1 || r.Flen == -1) && (l.Flen != r.Flen) && !mysql.IsIntegerType(l.Tp) {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("字段长度不一致(FieldType.Flen), 期望值: %v, 实际值: %v; ", l.Flen, r.Flen),
		}
	}

	if !(l.Decimal == -1 || r.Decimal == -1) && l.Decimal != r.Decimal {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("字段 Decimal 值不一致(FieldType.Decimal), 期望值: %v, 实际值: %v; ", l.Decimal, r.Decimal),
		}
	}

	if l.Collate != r.Collate {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("字段 FieldType.Collate 不一致, 期望值: %s, 实际值: %s; ", l.Collate, r.Collate),
		}
	}

	if mysql.HasUnsignedFlag(l.Flag) != mysql.HasUnsignedFlag(r.Flag) {
		return &Equal{
			equal: false,
			reason: fmt.Sprintf("字段 Unsigned 符号不一致, 期望值: %v, 实际值: %v; ",
				mysql.HasUnsignedFlag(l.Flag), mysql.HasUnsignedFlag(r.Flag)),
		}
	}

	if len(l.Elems) != len(r.Elems) {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("字段枚举值数量不一致, 期望值: %v, 实际值: %v; ", l.Elems, r.Elems),
		}
	}

	sort.Strings(l.Elems)
	sort.Strings(r.Elems)
	for i := range l.Elems {
		if l.Elems[i] != r.Elems[i] {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("字段枚举值[%v] 不一致, 期望值: %s, 实际值: %s; ", i, l.Elems[i], r.Elems[i]),
			}
		}
	}

	return &Equal{equal: true}
}
