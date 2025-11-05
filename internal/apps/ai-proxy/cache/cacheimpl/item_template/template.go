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

package item_template

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// templateCacheItem implements CacheItem for models
type templateCacheItem struct {
	*cachetypes.BaseCacheItem

	templatesByType templatetypes.TemplatesByType

	// for easy-use, currently templatesByType is immutable
	total uint64
	items []*templatetypes.TypeNamedTemplate
}

func NewTemplateCacheItem(dao dao.DAO, config *config.Config, templatesByType templatetypes.TemplatesByType) cachetypes.CacheItem {
	item := &templateCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypeTemplate, dao, config, item)
	item.templatesByType = templatesByType

	for typ, typedTemplates := range item.templatesByType {
		item.total += uint64(len(typedTemplates))
		for name, tpl := range typedTemplates {
			newTpl := templatetypes.TypeNamedTemplate{
				Type: typ,
				Name: name,
				Tpl:  tpl,
			}
			item.items = append(item.items, &newTpl)
		}
	}
	return item
}

func (c *templateCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	return c.total, c.items, nil
}

func (c *templateCacheItem) GetIDValue(item any) (string, error) {
	tpl, ok := item.(*templatetypes.TypeNamedTemplate)
	if !ok {
		return "", fmt.Errorf("invalid template data type")
	}
	return ConstructID(tpl.Type, tpl.Name), nil
}

func ConstructID(templateType templatetypes.TemplateType, templateName string) string {
	return fmt.Sprintf("%s:%s", string(templateType), templateName)
}
