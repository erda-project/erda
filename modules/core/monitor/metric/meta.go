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

package metric

import (
	"sort"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"google.golang.org/protobuf/types/known/structpb"
)

// TagsKeys .
func TagsKeys(m *pb.MetricMeta) []string {
	var keys []string
	for k := range m.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FieldsKeys .
func FieldsKeys(m *pb.MetricMeta) []string {
	var keys []string
	for k := range m.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// NewMeta .
func NewMeta() *pb.MetricMeta {
	return &pb.MetricMeta{
		Name:   &pb.NameDefine{},
		Tags:   make(map[string]*pb.TagDefine),
		Fields: make(map[string]*pb.FieldDefine),
	}
}

// CopyNameDefine .
func CopyNameDefine(name *pb.NameDefine) *pb.NameDefine {
	return &pb.NameDefine{
		Key:  name.Key,
		Name: name.Name,
	}
}

// CopyTagDefine .
func CopyTagDefine(tag *pb.TagDefine) *pb.TagDefine {
	clone := &pb.TagDefine{
		Key:  tag.Key,
		Name: tag.Name,
	}
	var values []*pb.ValueDefine
	for _, v := range tag.Values {
		values = append(values, CopyValue(v))
	}
	clone.Values = values
	return clone
}

// CopyTagDefine .
func CopyFieldDefine(field *pb.FieldDefine) *pb.FieldDefine {
	clone := &pb.FieldDefine{
		Key:  field.Key,
		Name: field.Name,
		Type: field.Type,
		Unit: field.Unit,
	}
	var values []*pb.ValueDefine
	for _, v := range field.Values {
		values = append(values, CopyValue(v))
	}
	clone.Values = values
	return clone
}

// CopyValue .
func CopyValue(v *pb.ValueDefine) *pb.ValueDefine {
	val := &pb.ValueDefine{
		Name: v.Name,
	}
	if v.Value != nil {
		value, _ := structpb.NewValue(v.Value.AsInterface())
		val.Value = value
	}
	return val
}

// CopyMeta
func CopyMeta(m *pb.MetricMeta) *pb.MetricMeta {
	n := &pb.MetricMeta{
		Tags:   make(map[string]*pb.TagDefine),
		Fields: make(map[string]*pb.FieldDefine),
	}
	n.Name = m.Name
	for k, t := range m.Tags {
		n.Tags[k] = CopyTagDefine(t)
	}
	for k, f := range m.Fields {
		n.Fields[k] = CopyFieldDefine(f)
	}
	return n
}
