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
			reason: "FieldType.Tp is not equal",
		}
	}

	if !(l.Flen == -1 || r.Flen == -1) && l.Flen != r.Flen {
		return &Equal{
			equal:  false,
			reason: "FieldType.Flen is not equal",
		}
	}

	if !(l.Decimal == -1 || r.Decimal == -1) && l.Decimal != r.Decimal {
		return &Equal{
			equal:  false,
			reason: "FieldType.Decimal is not equal",
		}
	}

	if l.Collate != r.Collate {
		return &Equal{
			equal:  false,
			reason: "FieldType.Collate is not equal",
		}
	}

	if mysql.HasUnsignedFlag(l.Flag) != mysql.HasUnsignedFlag(r.Flag) {
		return &Equal{
			equal:  false,
			reason: "FieldType's HasUnsignedFlag is not equal",
		}
	}

	if len(l.Elems) != len(r.Elems) {
		return &Equal{
			equal:  false,
			reason: "FieldType.Elems length is not equal",
		}
	}

	for i := range l.Elems {
		if l.Elems[i] != r.Elems[i] {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("FieldType.Elems[%v] is not equal", i),
			}
		}
	}

	return &Equal{equal: true}
}
