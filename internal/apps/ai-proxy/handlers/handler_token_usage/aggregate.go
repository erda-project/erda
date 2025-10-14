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
	"google.golang.org/protobuf/proto"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type tokenPricing struct {
	Currency string
	Input    float64
	Output   float64
	Request  float64
}

func (h *TokenUsageHandler) aggregateTokenUsages(ctx context.Context, records []*usagepb.TokenUsage, locale string) (*usagepb.TokenUsageAggregateResponse, error) {
	pricingCache := make(map[string]tokenPricing)
	var currency string
	totalCost := decimal.Zero

	resp := &usagepb.TokenUsageAggregateResponse{RecordCount: uint64(len(records))}

	for _, usage := range records {
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

		price := calculateUsagePrice(usage.InputTokens, usage.OutputTokens, pricing)
		priceDec := decimal.NewFromFloat(price)

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

		detailCost := priceDec.Round(4)
		resp.Details = append(resp.Details, &usagepb.TokenUsageDetail{
			Cost:         detailCost.InexactFloat64(),
			Currency:     pricing.Currency,
			RecordId:     usage.Id,
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			TotalTokens:  usage.TotalTokens,
			ModelId:      usage.ModelId,
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
	cachedModel, err := h.Cache.GetByID(ctx, cachetypes.ItemTypeModel, modelID)
	if err != nil {
		return nil, err
	}
	model, ok := cachedModel.(*modelpb.Model)
	if !ok || model == nil {
		return nil, fmt.Errorf("model %s not found", modelID)
	}

	cloned, ok := proto.Clone(model).(*modelpb.Model)
	if !ok {
		return nil, fmt.Errorf("failed to clone model %s", modelID)
	}

	if locale != "" && h.DAO != nil {
		if enhancer := i18n_services.NewMetadataEnhancerService(ctx, h.DAO); enhancer != nil {
			cloned = enhancer.EnhanceModelMetadata(ctx, cloned, locale)
		}
	}

	return cloned, nil
}

func calculateUsagePrice(inputTokens, outputTokens uint64, pricing tokenPricing) float64 {
	price := pricing.Request
	if pricing.Input != 0 && inputTokens > 0 {
		price += float64(inputTokens) * pricing.Input
	}
	if pricing.Output != 0 && outputTokens > 0 {
		price += float64(outputTokens) * pricing.Output
	}
	return price
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

	return tokenPricing{
		Currency: currency,
		Input:    promptPrice,
		Output:   completionPrice,
		Request:  requestPrice,
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
