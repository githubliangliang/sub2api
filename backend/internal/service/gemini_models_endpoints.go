package service

import (
	"context"
	"fmt"
	"sort"
)

// DeclaredModelIDs uses the same effective mapping keys as /v1/models. A single
// unmapped schedulable Gemini account keeps the existing upstream-list behavior.
func (s *GeminiMessagesCompatService) DeclaredModelIDs(ctx context.Context, groupID *int64) ([]string, error) {
	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, PlatformGemini, true)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	seen := make(map[string]struct{})
	for i := range accounts {
		if accounts[i].Platform != PlatformGemini {
			continue
		}
		mapping := accounts[i].GetModelMapping()
		if len(mapping) == 0 {
			return nil, nil
		}
		for model := range mapping {
			seen[model] = struct{}{}
		}
	}
	models := make([]string, 0, len(seen))
	for model := range seen {
		models = append(models, model)
	}
	sort.Strings(models)
	return models, nil
}

func DeclaredModelsCover(declared []string, model string) bool {
	normalized := normalizeRequestedModelForLookup(PlatformGemini, model)
	for _, pattern := range declared {
		if matchWildcard(pattern, model) || matchWildcard(pattern, normalized) {
			return true
		}
	}
	return false
}

// AntigravityGeminiModelIDs includes only models served by the existing mixed
// scheduling contract; forced/composite routing keeps its original handler path.
func (s *GeminiMessagesCompatService) AntigravityGeminiModelIDs(ctx context.Context, groupID *int64, requireMixed bool) ([]string, error) {
	accounts, err := s.listSchedulableAccountsOnce(ctx, groupID, PlatformAntigravity, false)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, account := range accounts {
		if account.Platform != PlatformAntigravity || (requireMixed && !account.IsMixedSchedulingEnabled()) {
			continue
		}
		for model := range account.GetModelMapping() {
			if isAntigravityGeminiModel(model) {
				seen[model] = struct{}{}
			}
		}
	}
	models := make([]string, 0, len(seen))
	for model := range seen {
		models = append(models, model)
	}
	sort.Strings(models)
	return models, nil
}
