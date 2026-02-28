package handler

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// modelFile is the default path used for the model list.
// You can override it with MODEL_FILE (useful for container deployments).
const modelFile = "model.json"

type modelFileSchema struct {
	Models []string `json:"models"`
}

// loadModelIDs loads the model id list from model.json.
func loadModelIDs() ([]string, error) {
	path := os.Getenv("MODEL_FILE")
	if strings.TrimSpace(path) == "" {
		path = modelFile
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s modelFileSchema
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(s.Models))
	seen := make(map[string]struct{}, len(s.Models))
	for _, m := range s.Models {
		id := strings.TrimSpace(m)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("no models found in model.json")
	}
	return ids, nil
}

// inferOwner provides a best-effort owned_by value for /v1/models responses.
// It's purely cosmetic for compatibility and does not affect routing/selection.
func inferOwner(modelID string) string {
	id := strings.TrimSpace(modelID)
	if id == "" {
		return "unknown"
	}
	if strings.HasPrefix(id, "openrouter:") {
		return "openrouter"
	}
	if strings.HasPrefix(id, "togetherai:") {
		return "togetherai"
	}
	if strings.HasPrefix(id, "anthropic:") || strings.Contains(id, "claude") {
		return "anthropic"
	}

	// Default for common OpenAI-family ids.
	if strings.HasPrefix(id, "gpt-") || strings.HasPrefix(id, "o1") || strings.HasPrefix(id, "o3") || strings.HasPrefix(id, "o4") {
		return "openai"
	}

	// Heuristic: prefix before the first ':' is usually a provider tag.
	if i := strings.IndexByte(id, ':'); i > 0 {
		return id[:i]
	}
	return "unknown"
}
