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

package i18n_services

import (
	"context"
	"strings"
	"sync"

	"golang.org/x/text/language"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	infrai18n "github.com/erda-project/erda-infra/providers/i18n"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// I18nConfigCache i18n configuration cache
type I18nConfigCache struct {
	// key: category_itemKey_fieldName_locale, value: i18n configuration
	configs map[string]*i18n.Config
	mu      sync.RWMutex
}

// MetadataEnhancerService metadata enhancement service
type MetadataEnhancerService struct {
	dao   dao.DAO
	cache *I18nConfigCache
}

// NewMetadataEnhancerService create metadata enhancement service
func NewMetadataEnhancerService(ctx context.Context, dao dao.DAO) *MetadataEnhancerService {
	s := &MetadataEnhancerService{
		dao: dao,
		cache: &I18nConfigCache{
			configs: make(map[string]*i18n.Config),
		},
	}
	if err := s.PreloadI18nConfigs(); err != nil {
		ctxhelper.MustGetLogger(ctx).Warnf("failed to preload i18n configs: %v", err)
	}
	return s
}

// PreloadI18nConfigs preload all i18n configurations
func (s *MetadataEnhancerService) PreloadI18nConfigs() error {
	// get all i18n configurations
	allConfigs, err := s.dao.I18nClient().GetAllConfigs()
	if err != nil {
		return err
	}

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	// build cache mapping
	for _, config := range allConfigs {
		key := i18n.BuildConfigKey(config.Category, config.ItemKey, config.FieldName, config.Locale)
		s.cache.configs[key] = config
	}

	return nil
}

// getConfigFromCache get configuration from cache
func (s *MetadataEnhancerService) getConfigFromCache(category, itemKey, fieldName, locale string) (*i18n.Config, bool) {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	key := i18n.BuildConfigKey(category, itemKey, fieldName, locale)
	config, exists := s.cache.configs[key]
	return config, exists
}

// EnhanceModelMetadata enhance model metadata information
func (s *MetadataEnhancerService) EnhanceModelMetadata(ctx context.Context, model *modelpb.Model, locale string) *modelpb.Model {
	if model == nil || model.Metadata == nil || model.Metadata.Public == nil {
		return model
	}

	// parse to metadata.Metadata for convenience
	meta := metadata.FromProtobuf(model.Metadata)

	// enhance publisher info
	s.enhancePublisherInfo(model.Publisher, meta, locale)

	// enhance abilities info
	s.enhanceAbilitiesInfo(meta, locale)

	// set logo
	s.setModelLogo(model.Publisher, meta)

	// enhance pricing info
	s.enhancePricingInfo(meta, locale)

	model.Metadata = meta.ToProtobuf()

	return model
}

// enhancePublisherInfo enhance publisher information
func (s *MetadataEnhancerService) enhancePublisherInfo(publisher string, meta metadata.Metadata, locale string) {
	if publisher == "" {
		return
	}
	// insert publisher_${locale}
	publisherLocaleKey := "publisher_" + locale
	publisherLocaleValue := publisher
	if config, ok := s.getConfigFromCache(string(i18n.CategoryPublisher), publisher, string(i18n.FieldNameValue), locale); ok && config.Value != "" {
		publisherLocaleValue = config.Value
	}
	meta.Public[publisherLocaleKey] = publisherLocaleValue
	meta.Public["publisher_i18n"] = publisherLocaleValue
}

// enhanceAbilitiesInfo enhance abilities information
func (s *MetadataEnhancerService) enhanceAbilitiesInfo(meta metadata.Metadata, locale string) {
	abilitiesObjValue := meta.Public["abilities"]
	if abilitiesObjValue == nil {
		return
	}
	abilitiesObj, ok := abilitiesObjValue.(map[string]any)
	if !ok {
		return
	}
	abilitiesV, ok := abilitiesObj["abilities"]
	if !ok {
		return
	}
	var abilities []string
	cputil.MustObjJSONTransfer(abilitiesV, &abilities)

	var localeAbilities []string
	for _, ability := range abilities {
		ability := ability
		if config, ok := s.getConfigFromCache(string(i18n.CategoryAbility), ability, string(i18n.FieldNameValue), locale); ok && config.Value != "" {
			localeAbilities = append(localeAbilities, config.Value)
		} else {
			localeAbilities = append(localeAbilities, ability)
		}
	}

	localeAbilitiesKey := "abilities_" + locale
	meta.Public["abilities"].(map[string]any)[localeAbilitiesKey] = localeAbilities
	meta.Public["abilities"].(map[string]any)["abilities_i18n"] = localeAbilities
}

// setModelLogo set model logo
func (s *MetadataEnhancerService) setModelLogo(publisher string, meta metadata.Metadata) {
	if publisher == "" {
		return
	}
	config, ok := s.getConfigFromCache(string(i18n.CategoryPublisher), publisher, string(i18n.FieldLogo), string(i18n.LocaleUniversal))
	if !ok {
		return
	}
	meta.Public["logo"] = config.Value
}

// GetLocaleFromContext get locale from request context or header
func GetLocaleFromContext(inputLang string) string {
	// get from header
	langs, err := infrai18n.ParseLanguageCode(inputLang)
	if err != nil {
		return string(i18n.LocaleDefault)
	}
	if len(langs) == 0 {
		return string(i18n.LocaleDefault)
	}

	// check valid
	lang := langs[0]
	tag, err := language.Parse(lang.Code)
	if err != nil {
		return string(i18n.LocaleDefault)
	}
	// trim -
	return strings.Split(tag.String(), "-")[0]
}
