// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
			reason: fmt.Sprintf("FieldType.Tp is not equal, expected: %s, actual: %s; ", l.String(), r.String()),
		}
	}

	if !(l.Flen == -1 || r.Flen == -1) && (l.Flen != r.Flen) && !mysql.IsIntegerType(l.Tp) {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("FieldType.Flen is not equal, expected: %v, actual: %v; ", l.Flen, r.Flen),
		}
	}

	if !(l.Decimal == -1 || r.Decimal == -1) && l.Decimal != r.Decimal {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("FieldType.Decimal is not equal, expected: %v, actual: %v; ", l.Decimal, r.Decimal),
		}
	}

	if l.Collate != r.Collate {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("FieldType.Collate is not equal, expected: %s, actual: %s; ", l.Collate, r.Collate),
		}
	}

	if mysql.HasUnsignedFlag(l.Flag) != mysql.HasUnsignedFlag(r.Flag) {
		return &Equal{
			equal: false,
			reason: fmt.Sprintf("FieldType's HasUnsignedFlag is not equal, expected unsigned flag: %v, actual unsigned flag: %v; ",
				mysql.HasUnsignedFlag(l.Flag), mysql.HasUnsignedFlag(r.Flag)),
		}
	}

	if len(l.Elems) != len(r.Elems) {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("FieldType.Elems length is not equal, expected: %v, actual: %v; ", l.Elems, r.Elems),
		}
	}

	sort.Strings(l.Elems)
	sort.Strings(r.Elems)
	for i := range l.Elems {
		if l.Elems[i] != r.Elems[i] {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("FieldType.Elems[%v] is not equal, expected: %s, actual: %s; ", i, l.Elems[i], r.Elems[i]),
			}
		}
	}

	return &Equal{equal: true}
}
