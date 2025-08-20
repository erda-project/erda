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

package logutil

import (
	"context"
	"reflect"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

// InjectLoggerWithFilterInfo injects a sub-logger with filter info into the context
func InjectLoggerWithFilterInfo[T any](ctx context.Context, filter filter_define.FilterWithName[T]) {
	// Use reflection to get package name from filter type safely
	t := reflect.TypeOf(filter.Instance)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var pkgPath string
	if t != nil {
		pkgPath = t.PkgPath()
	}

	// Extract two levels of package path (e.g., "filters/after-audit" from "github.com/.../filters/after-audit")
	parts := strings.Split(pkgPath, "/")
	var packageName string
	if len(parts) >= 2 {
		packageName = parts[len(parts)-2] + "/" + parts[len(parts)-1]
	} else if pkgPath != "" {
		if i := strings.LastIndex(pkgPath, "/"); i >= 0 {
			packageName = pkgPath[i+1:]
		} else {
			packageName = pkgPath
		}
	} else {
		packageName = "unknown"
	}

	// Add filter name
	fullFilterPath := packageName + "@" + filter.Name

	// Stage - determine stage from generic type T
	var stage string
	switch any(filter.Instance).(type) {
	case filter_define.ProxyRequestRewriter:
		stage = "request"
	case filter_define.ProxyResponseModifier:
		stage = "response"
	default:
		stage = "unknown"
	}

	logger := ctxhelper.MustGetLogger(ctx)
	logger.Set("filter", fullFilterPath)
	logger.Set("stage", stage)
	ctxhelper.PutLogger(ctx, logger)
}
