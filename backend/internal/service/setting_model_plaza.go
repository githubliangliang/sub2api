package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ModelPlazaReferencePricing is a display-only price in USD per token.
// Pointers distinguish an explicit zero from a missing price.
type ModelPlazaReferencePricing struct {
	InputPrice        *float64 `json:"input_price"`
	OutputPrice       *float64 `json:"output_price"`
	CacheWritePrice   *float64 `json:"cache_write_price"`
	CacheWrite1hPrice *float64 `json:"cache_write_1h_price"`
	CacheReadPrice    *float64 `json:"cache_read_price"`
}

func modelPlazaReferencePricingKey(platform, model string) string {
	// Keep keys within settings.key's 100-character limit, including namespaced IDs.
	return fmt.Sprintf("model_plaza_price_%x", sha256.Sum256([]byte(platform+"\x00"+model)))
}

func (p *ModelPlazaReferencePricing) validate() error {
	if p.InputPrice == nil || p.OutputPrice == nil {
		return infraerrors.BadRequest("INVALID_REFERENCE_PRICE", "Input and output prices are required")
	}
	for _, price := range []*float64{p.InputPrice, p.OutputPrice, p.CacheWritePrice, p.CacheWrite1hPrice, p.CacheReadPrice} {
		if price != nil && (*price < 0 || math.IsNaN(*price) || math.IsInf(*price, 0)) {
			return infraerrors.BadRequest("INVALID_REFERENCE_PRICE", "Prices must be finite, nonnegative numbers")
		}
	}
	return nil
}

// SetModelPlazaReferencePricing stores one platform/model independently so edits
// to different rows cannot overwrite each other. A nil price removes the supplement.
func (s *SettingService) SetModelPlazaReferencePricing(ctx context.Context, platform, model string, price *ModelPlazaReferencePricing) error {
	if strings.TrimSpace(platform) == "" || strings.TrimSpace(model) == "" || strings.ContainsRune(platform, '\x00') {
		return infraerrors.BadRequest("INVALID_REFERENCE_PRICE", "Platform and client model ID are required")
	}
	key := modelPlazaReferencePricingKey(platform, model)
	if price == nil {
		return s.settingRepo.Delete(ctx, key)
	}
	if err := price.validate(); err != nil {
		return err
	}
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, key, string(data))
}

// ApplyModelPlazaReferencePricing reads supplements only for the admin inventory.
// It never modifies the shared pricing catalog or any billing configuration.
func (s *SettingService) ApplyModelPlazaReferencePricing(ctx context.Context, items []AccountModelPlazaItem) error {
	if s == nil || len(items) == 0 {
		return nil
	}
	keys := make([]string, len(items))
	for i := range items {
		keys[i] = modelPlazaReferencePricingKey(items[i].Platform, items[i].ClientID)
	}
	values, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return err
	}
	for i, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		var price ModelPlazaReferencePricing
		if err := json.Unmarshal([]byte(value), &price); err != nil {
			return fmt.Errorf("decode model plaza reference price: %w", err)
		}
		if err := price.validate(); err != nil {
			return fmt.Errorf("invalid stored model plaza reference price: %w", err)
		}
		items[i].Pricing = &AccountModelPlazaPricing{
			Model:             items[i].ClientID,
			Manual:            true,
			InputPrice:        *price.InputPrice,
			OutputPrice:       *price.OutputPrice,
			CacheWritePrice:   price.CacheWritePrice,
			CacheWrite1hPrice: price.CacheWrite1hPrice,
			CacheReadPrice:    price.CacheReadPrice,
		}
	}
	return nil
}
