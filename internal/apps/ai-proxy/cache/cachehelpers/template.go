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

package cachehelpers

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
)

func GetTemplateByTypeName(ctx context.Context, typ templatetypes.TemplateType, name string, placeholderValues map[string]string) (*pb.Template, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	templateV, err := cache.GetByID(ctx, cachetypes.ItemTypeTemplate, item_template.ConstructID(typ, name))
	if err != nil {
		return nil, err
	}
	tpl := templateV.(*item_template.TypeNamedTemplate).Tpl
	if tpl.GetDeprecated() {
		return nil, fmt.Errorf("template %s of type %s is deprecated", name, typ)
	}
	// check template
	if err := checkTemplate(tpl, placeholderValues); err != nil {
		return nil, fmt.Errorf("failed to check template: %w", err)
	}
	return tpl, nil
}

func checkTemplate(tpl *pb.Template, placeholderValues map[string]string) error {
	for _, placeholder := range tpl.Placeholders {
		if !placeholder.Required {
			continue
		}
		placeholderValue, ok := placeholderValues[placeholder.Name]
		if !ok {
			return fmt.Errorf("placeholder %q not found", placeholder.Name)
		}
		// todo check type
		_ = placeholderValue
	}
	return nil
}
