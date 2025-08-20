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

package notes

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type DBWriter struct {
	db dao.DAO
}

func NewDBWriter(in dao.DAO) *DBWriter {
	return &DBWriter{db: in}
}

var table = audit.Audit{}
var tableFieldMap = map[string]string{}

func init() {
	// use reflect to get tableFieldMap from gorm:column tags
	t := reflect.TypeOf(table)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			// extract column name from gorm tag
			parts := strings.Split(gormTag, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "column:") {
					columnName := strings.TrimPrefix(part, "column:")
					tableFieldMap[columnName] = field.Name
					break
				}
			}
		}
	}
}

func (w *DBWriter) Write(ctx context.Context, p types.Patch) {
	rec := &audit.Audit{BaseModel: common.BaseModelWithID(p.AuditID)}
	if err := w.db.AuditClient().DB.Model(rec).First(rec).Error; err != nil {
		ctxhelper.MustGetLoggerBase(ctx).Warnf("failed to get audit record from db: %v", err)
		return
	}
	if rec.Metadata.Public == nil {
		rec.Metadata.Public = map[string]any{}
	}

	for k, v := range p.Notes {
		// map column name -> struct field name
		fieldName, exists := tableFieldMap[k]
		if !exists {
			// not a direct column: store under metadata.public
			rec.Metadata.Public[k] = v
			continue
		}

		// Build a tiny updates map for this single field
		updates := map[string]any{fieldName: v}

		// Decoder with weak typing + custom hooks
		cfg := &mapstructure.DecoderConfig{
			Result:           rec,
			WeaklyTypedInput: true,
			ZeroFields:       false, // don't zero other fields
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				// time.Time: RFC3339 / common formats / unix seconds or milliseconds
				func(from, to reflect.Type, data any) (any, error) {
					if to != reflect.TypeOf(time.Time{}) {
						return data, nil
					}
					switch x := data.(type) {
					case time.Time:
						return x, nil
					case string:
						s := strings.TrimSpace(x)
						formats := []string{
							time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02",
						}
						for _, f := range formats {
							if t, err := time.Parse(f, s); err == nil {
								return t, nil
							}
						}
						if i, err := strconv.ParseInt(s, 10, 64); err == nil {
							if i > 1e12 {
								return time.UnixMilli(i), nil
							}
							return time.Unix(i, 0), nil
						}
					case int64:
						return time.Unix(x, 0), nil
					case int:
						return time.Unix(int64(x), 0), nil
					case float64:
						return time.Unix(int64(x), 0), nil
					}
					return data, nil
				},
				// []byte: allow string -> []byte
				func(from, to reflect.Type, data any) (any, error) {
					if to.Kind() == reflect.Slice && to.Elem().Kind() == reflect.Uint8 {
						switch s := data.(type) {
						case string:
							return []byte(s), nil
						}
					}
					return data, nil
				},
				// JSON string -> struct/map/slice
				func(from, to reflect.Type, data any) (any, error) {
					s, ok := data.(string)
					if !ok {
						return data, nil
					}
					dstKind := to.Kind()
					if !(dstKind == reflect.Struct || dstKind == reflect.Map || dstKind == reflect.Slice) {
						return data, nil
					}
					trim := strings.TrimSpace(s)
					if len(trim) == 0 || (trim[0] != '{' && trim[0] != '[' && trim != "null") {
						return data, nil
					}
					dst := reflect.New(to).Interface()
					if err := json.Unmarshal([]byte(s), dst); err != nil {
						return data, nil
					}
					return reflect.ValueOf(dst).Elem().Interface(), nil
				},
			),
		}
		dec, err := mapstructure.NewDecoder(cfg)
		if err != nil {
			ctxhelper.MustGetLoggerBase(ctx).Warnf("failed to create decoder: %v", err)
			rec.Metadata.Public[k] = v
			continue
		}
		if err := dec.Decode(updates); err != nil {
			// decoding for this field failed; fall back to metadata
			ctxhelper.MustGetLoggerBase(ctx).Warnf("decode field %s (col=%s) failed: %v; fallback to metadata", fieldName, k, err)
			rec.Metadata.Public[k] = v
			continue
		}
	}
	if _, err := w.db.AuditClient().Update(ctx, rec); err != nil {
		ctxhelper.MustGetLoggerBase(ctx).Warnf("failed to update audit record: %v", err)
		return
	}
}
