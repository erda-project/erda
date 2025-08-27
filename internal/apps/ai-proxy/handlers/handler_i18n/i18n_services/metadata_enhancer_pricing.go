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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

// enhancePricingInfo enhance pricing information with currency conversion based on locale
//
// "pricing":
//
//	{
//	  "completion": "0.00001",
//	  "image": "0.003613",
//	  "input_cache_read": "0.00000125",
//	  "input_cache_write": "0",
//	  "internal_reasoning": "0",
//	  "prompt": "0.0000025",
//	  "request": "0",
//	  "unit": "USD",
//	  "web_search": "0"
//	}
func (s *MetadataEnhancerService) enhancePricingInfo(meta metadata.Metadata, locale string) {
	pricingV := meta.Public["pricing"]
	if pricingV == nil {
		return
	}
	pricing, ok := pricingV.(map[string]any)
	if !ok {
		return
	}
	// get unit
	unitV, ok := pricing["unit"]
	if !ok {
		return
	}
	unit, ok := unitV.(string)
	if !ok {
		return
	}

	// fix unit: RMB -> CNY
	if unit == "RMB" {
		pricing["unit"] = i18n.CurrencyCNY
	}

	currentCurrency := parseUnit(unit)
	targetCurrency := getTargetUnit(locale)
	if currentCurrency == targetCurrency {
		return
	}

	// do convert
	exchangeRate := s.getExchangeRate(currentCurrency, targetCurrency)
	if exchangeRate <= 0 {
		return
	}
	convertPricingValues(pricing, exchangeRate)
	pricing["unit"] = string(targetCurrency)
}

// getExchangeRate get exchange rate with fallback strategy
// Priority: Real-time API > i18n configuration > Fixed rates
func (s *MetadataEnhancerService) getExchangeRate(currentCurrency, targetCurrency i18n.Currency) float64 {
	// 1. try to get from real-time API
	if rate := getExchangeRateFromCache(currentCurrency, targetCurrency); rate > 0 {
		return rate
	}

	// 2. try to get from i18n configuration
	exchangeRateKey := generateExchangeRateKey(currentCurrency, targetCurrency)
	if config, ok := s.getConfigFromCache(string(i18n.CategoryPricing), string(i18n.FieldExchangeRate), exchangeRateKey, string(i18n.LocaleUniversal)); ok && config.Value != "" {
		if rate, err := strconv.ParseFloat(config.Value, 64); err == nil && rate > 0 {
			return rate
		}
	}

	return InvalidExchangeRate
}

const InvalidExchangeRate = -1.0

var (
	exchangeRateMap  = map[string]float64{}
	currencyRateLock sync.Mutex
)

func generateExchangeRateKey(currentCurrency, targetCurrency i18n.Currency) string {
	return fmt.Sprintf("%s_TO_%s", currentCurrency, targetCurrency)
}

func init() {
	go func() {
		for {
			preloadAllExchangeRates()
			time.Sleep(time.Hour * 24)
		}
	}()
}

// preloadAllExchangeRates preload all supported exchange rates
func preloadAllExchangeRates() {
	_preloadExchangeRate(i18n.CurrencyUSD, i18n.CurrencyCNY)
	_preloadExchangeRate(i18n.CurrencyCNY, i18n.CurrencyUSD)
}

func _preloadExchangeRate(currentCurrency, targetCurrency i18n.Currency) {
	currencyRateLock.Lock()
	defer currencyRateLock.Unlock()

	rate := fetchRealTimeExchangeRate(currentCurrency, targetCurrency)
	exchangeRateMap[generateExchangeRateKey(currentCurrency, targetCurrency)] = rate
}

func parseUnit(unit string) i18n.Currency {
	switch strings.ToUpper(unit) {
	case "CNY", "RMB":
		return i18n.CurrencyCNY
	case "USD":
		return i18n.CurrencyUSD
	default:
		return i18n.CurrencyDefault
	}
}

// getTargetUnit get target currency unit based on locale
func getTargetUnit(locale string) i18n.Currency {
	switch i18n.Locale(locale) {
	case i18n.LocaleChinese:
		return i18n.CurrencyCNY
	case i18n.LocaleEnglish:
		return i18n.CurrencyUSD
	default:
		return i18n.CurrencyDefault
	}
}

func getExchangeRateFromCache(currentCurrency, targetCurrency i18n.Currency) float64 {
	currencyRateLock.Lock()
	rate, ok := exchangeRateMap[generateExchangeRateKey(currentCurrency, targetCurrency)]
	currencyRateLock.Unlock()
	if ok {
		return rate
	}
	// if not fetched yet, trigger once fetch
	_preloadExchangeRate(currentCurrency, targetCurrency)
	return InvalidExchangeRate
}

// convertPricingValues convert all pricing values by given rate
func convertPricingValues(pricing map[string]any, rate float64) {
	for key, value := range pricing {
		if key == "unit" {
			continue
		}
		valueS, ok := value.(string)
		if !ok {
			continue
		}
		if price, err := strconv.ParseFloat(valueS, 64); err == nil {
			convertedPrice := price * rate
			pricing[key] = formatPriceWithSignificantDigits(convertedPrice, 3)
		}
	}
}

// formatPriceWithSignificantDigits formats price with specified significant digits in decimal format
func formatPriceWithSignificantDigits(value float64, digits int) string {
	if value == 0 {
		return "0"
	}

	// first get the result with %g format to handle significant digits
	formatted := fmt.Sprintf("%.*g", digits, value)

	// if it's in scientific notation, convert back to decimal
	if strings.Contains(formatted, "e") {
		// parse the float and format with enough decimal places
		if f, err := strconv.ParseFloat(formatted, 64); err == nil {
			// calculate needed decimal places
			str := fmt.Sprintf("%.15f", f)
			// trim trailing zeros
			str = strings.TrimRight(str, "0")
			str = strings.TrimRight(str, ".")
			return str
		}
	}

	return formatted
}

func fetchRealTimeExchangeRate(currentCurrency, targetCurrency i18n.Currency) (result float64) {
	defer func() {
		if r := recover(); r != nil {
			result = -1
		}
	}()

	req, err := http.NewRequest(http.MethodGet, "https://open.er-api.com/v6/latest/"+string(currentCurrency), nil)
	if err != nil {
		return InvalidExchangeRate
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return InvalidExchangeRate
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return InvalidExchangeRate
	}

	m := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return InvalidExchangeRate
	}

	// extract rate
	rates, ok := m["rates"].(map[string]any)
	if !ok {
		return InvalidExchangeRate
	}

	rateValue, ok := rates[string(targetCurrency)]
	if !ok {
		return InvalidExchangeRate
	}

	// handle both float64 and int cases
	switch v := rateValue.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return InvalidExchangeRate
	}
}
