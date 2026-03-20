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

package handler_token_usage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type tokenPricing struct {
	Currency          string
	Input             float64
	Output            float64
	Request           float64
	InputCacheRead    float64
	HasInputCacheRead bool
}

func (h *TokenUsageHandler) aggregateTokenUsages(ctx context.Context, records []*usagepb.TokenUsage, locale string) (*usagepb.TokenUsageAggregateResponse, error) {
	pricingCache := make(map[string]tokenPricing)
	var currency string
	totalCost := decimal.Zero

	resp := &usagepb.TokenUsageAggregateResponse{RecordCount: uint64(len(records)), IsEstimated: false}

	for _, usage := range records {
		if usage.IsEstimated {
			resp.IsEstimated = true
		}

		if usage.TotalTokens == 0 {
			usage.TotalTokens = usage.InputTokens + usage.OutputTokens
		}

		resp.TotalInputTokens += usage.InputTokens
		resp.TotalOutputTokens += usage.OutputTokens
		resp.TotalTokens += usage.TotalTokens

		pricing, err := h.resolveModelPricing(ctx, usage.ModelId, locale, pricingCache)
		if err != nil {
			if logger, ok := ctxhelper.GetLogger(ctx); ok {
				logger.Errorf("failed to resolve pricing for model %s: %v", usage.ModelId, err)
			}
		}

		price, priceEstimated := calculateUsagePrice(usage, pricing)
		priceDec := decimal.NewFromFloat(price)
		isEstimated := usage.IsEstimated || priceEstimated

		if pricing.Currency != "" {
			if currency == "" {
				currency = pricing.Currency
			} else if currency != pricing.Currency {
				return nil, fmt.Errorf("mixed pricing currencies detected: %s vs %s", currency, pricing.Currency)
			}
		}

		if priceDec.Sign() > 0 {
			totalCost = totalCost.Add(priceDec)
		}

		if isEstimated {
			resp.IsEstimated = true
		}

		detailCost := priceDec.Round(4)
		resp.Details = append(resp.Details, &usagepb.TokenUsageDetail{
			Cost:         detailCost.InexactFloat64(),
			Currency:     pricing.Currency,
			RecordId:     usage.Id,
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			TotalTokens:  usage.TotalTokens,
			ModelId:      usage.ModelId,
			CreatedAt:    usage.CreatedAt,
			IsEstimated:  isEstimated,
		})
	}

	resp.TotalCost = totalCost.Round(4).InexactFloat64()
	resp.Currency = currency
	return resp, nil
}

func (h *TokenUsageHandler) resolveModelPricing(ctx context.Context, modelID, locale string, pricingCache map[string]tokenPricing) (tokenPricing, error) {
	if modelID == "" {
		return tokenPricing{}, nil
	}
	if pricing, ok := pricingCache[modelID]; ok {
		return pricing, nil
	}
	model, err := h.resolveModel(ctx, modelID, locale)
	if err != nil {
		pricingCache[modelID] = tokenPricing{}
		return tokenPricing{}, err
	}
	pricing := extractPricingFromModel(model)
	pricingCache[modelID] = pricing
	return pricing, nil
}

func (h *TokenUsageHandler) resolveModel(ctx context.Context, modelID, locale string) (*modelpb.Model, error) {
	if h.Cache == nil {
		return nil, fmt.Errorf("model cache is not configured")
	}
	model, err := cachehelpers.GetRenderedModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}
	if locale != "" && h.DAO != nil {
		if enhancer := i18n_services.NewMetadataEnhancerService(ctx, h.DAO); enhancer != nil {
			model = enhancer.EnhanceModelMetadata(ctx, model, locale)
		}
	}
	return model, nil
}

func calculateUsagePrice(usage *usagepb.TokenUsage, pricing tokenPricing) (float64, bool) {
	price := pricing.Request
	if usage == nil {
		return price, false
	}

	inputTokens := usage.InputTokens
	outputTokens := usage.OutputTokens
	cachedInputTokens, hasCachedInput := extractCachedInputTokens(usage.UsageDetails)
	if cachedInputTokens > inputTokens {
		cachedInputTokens = inputTokens
	}

	regularInputTokens := inputTokens
	if hasCachedInput && cachedInputTokens > 0 {
		regularInputTokens -= cachedInputTokens
	}

	if pricing.Input != 0 && regularInputTokens > 0 {
		price += float64(regularInputTokens) * pricing.Input
	}

	isEstimated := false
	if hasCachedInput && cachedInputTokens > 0 {
		cacheReadPrice := pricing.InputCacheRead
		if !pricing.HasInputCacheRead {
			cacheReadPrice = pricing.Input / 5
			isEstimated = true
		}
		if cacheReadPrice != 0 {
			price += float64(cachedInputTokens) * cacheReadPrice
		}
	}

	if pricing.Output != 0 && outputTokens > 0 {
		price += float64(outputTokens) * pricing.Output
	}
	return price, isEstimated
}

func extractPricingFromModel(model *modelpb.Model) tokenPricing {
	if model == nil || model.Metadata == nil {
		return tokenPricing{}
	}
	meta := metadata.FromProtobuf(model.Metadata)
	rawPricing, ok := meta.Public["pricing"]
	if !ok {
		return tokenPricing{}
	}
	pricingMap, ok := rawPricing.(map[string]any)
	if !ok {
		return tokenPricing{}
	}

	normalized := make(map[string]any, len(pricingMap))
	for k, v := range pricingMap {
		normalized[strings.ToLower(k)] = v
	}

	currency := stringValue(normalized["unit"])
	promptPrice, _ := parsePriceValue(normalized["prompt"])
	completionPrice, _ := parsePriceValue(normalized["completion"])
	requestPrice, _ := parsePriceValue(normalized["request"])
	cacheReadPrice, hasCacheReadPrice := parsePriceValue(normalized["input_cache_read"])

	return tokenPricing{
		Currency:          currency,
		Input:             promptPrice,
		Output:            completionPrice,
		Request:           requestPrice,
		InputCacheRead:    cacheReadPrice,
		HasInputCacheRead: hasCacheReadPrice,
	}
}

func extractCachedInputTokens(usageDetails string) (uint64, bool) {
	usageDetails = strings.TrimSpace(usageDetails)
	if usageDetails == "" || usageDetails == "{}" {
		return 0, false
	}

	var payload any
	if err := json.Unmarshal([]byte(usageDetails), &payload); err != nil {
		return 0, false
	}
	return findCachedInputTokens(payload)
}

func findCachedInputTokens(v any) (uint64, bool) {
	switch val := v.(type) {
	case map[string]any:
		if cached, ok := getNestedUint(val, "input_tokens_details", "cached_tokens"); ok {
			return cached, true
		}
		if cached, ok := getNestedUint(val, "prompt_tokens_details", "cached_tokens"); ok {
			return cached, true
		}
		for _, nested := range val {
			if cached, ok := findCachedInputTokens(nested); ok {
				return cached, true
			}
		}
	case []any:
		for _, item := range val {
			if cached, ok := findCachedInputTokens(item); ok {
				return cached, true
			}
		}
	}
	return 0, false
}

func getNestedUint(m map[string]any, key, nestedKey string) (uint64, bool) {
	if m == nil {
		return 0, false
	}
	nested, ok := m[key].(map[string]any)
	if !ok {
		return 0, false
	}
	return parseUintValue(nested[nestedKey])
}

func parseUintValue(v any) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case float32:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int32:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case uint:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	case json.Number:
		i, err := val.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	default:
		return 0, false
	}
}

func parsePriceValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		if v == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}
