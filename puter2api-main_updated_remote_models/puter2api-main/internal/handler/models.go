package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// modelFile is the default path used for the model list.
// You can override it with MODEL_FILE (useful for container deployments).
const modelFile = "model.json"

// Env toggles for model listing behavior.
//
// By default, we keep the original behavior: use local model.json.
// If MODELS_SOURCE=remote, /v1/models will fetch Puter model list and cache it.
const (
	envModelsSource    = "MODELS_SOURCE"      // "remote" to enable Puter fetch
	envModelsTTL       = "MODELS_TTL_SECONDS" // cache ttl in seconds (default 900)
	envModelsRemoteURL = "PUTER_MODELS_URL"   // override remote URL
)

// Default Puter models endpoint documented by Puter.
// It returns model details; we only extract the id list.
const defaultPuterModelsURL = "https://api.puter.com/puterai/chat/models/details"

var (
	remoteModelsMu     sync.Mutex
	remoteModelsCached []string
	remoteModelsExpiry time.Time
)

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

// modelsSourceRemoteEnabled returns true if the user explicitly requested remote model listing.
func modelsSourceRemoteEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(envModelsSource)), "remote")
}

func modelsCacheTTL() time.Duration {
	// Default 15 minutes.
	ttl := 15 * time.Minute
	v := strings.TrimSpace(os.Getenv(envModelsTTL))
	if v == "" {
		return ttl
	}
	sec, err := strconv.Atoi(v)
	if err != nil || sec <= 0 {
		return ttl
	}
	return time.Duration(sec) * time.Second
}

func puterModelsURL() string {
	u := strings.TrimSpace(os.Getenv(envModelsRemoteURL))
	if u == "" {
		return defaultPuterModelsURL
	}
	return u
}

// loadModelIDsRemoteCached fetches model ids from Puter and caches the result.
// It is best-effort: any error should be handled by falling back to local model.json.
func loadModelIDsRemoteCached(ctx context.Context) ([]string, error) {
	ttl := modelsCacheTTL()
	now := time.Now()

	remoteModelsMu.Lock()
	if len(remoteModelsCached) > 0 && now.Before(remoteModelsExpiry) {
		ids := append([]string(nil), remoteModelsCached...)
		remoteModelsMu.Unlock()
		return ids, nil
	}
	remoteModelsMu.Unlock()

	ids, err := fetchPuterModelIDs(ctx, puterModelsURL())
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("no models returned from remote")
	}

	remoteModelsMu.Lock()
	remoteModelsCached = append([]string(nil), ids...)
	remoteModelsExpiry = time.Now().Add(ttl)
	remoteModelsMu.Unlock()
	return ids, nil
}

func fetchPuterModelIDs(ctx context.Context, url string) ([]string, error) {
	// Keep this client small; /v1/models should be fast.
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "puter2api/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remote models request failed: status=%d", resp.StatusCode)
	}

	var payload any
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}

	return extractModelIDs(payload)
}

// extractModelIDs tries multiple common shapes to extract a model id list.
// We keep it lenient because the upstream payload is not part of our API contract.
func extractModelIDs(payload any) ([]string, error) {
	var items []any

	switch v := payload.(type) {
	case []any:
		items = v
	case map[string]any:
		// Common wrappers
		for _, key := range []string{"models", "data", "items", "result"} {
			if raw, ok := v[key]; ok {
				switch vv := raw.(type) {
				case []any:
					items = vv
				}
				if items != nil {
					break
				}
			}
		}
		if items == nil {
			// If it's a map without a wrapper, treat each value as a candidate record.
			items = make([]any, 0, len(v))
			for _, vv := range v {
				items = append(items, vv)
			}
		}
	default:
		return nil, errors.New("unexpected remote payload")
	}

	seen := map[string]struct{}{}
	ids := make([]string, 0, len(items))
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		ids = append(ids, s)
	}

	for _, it := range items {
		switch vv := it.(type) {
		case string:
			add(vv)
		case map[string]any:
			// Most likely field names
			for _, key := range []string{"id", "model", "name"} {
				if raw, ok := vv[key]; ok {
					if s, ok := raw.(string); ok {
						add(s)
						break
					}
				}
			}
		}
	}

	if len(ids) == 0 {
		return nil, errors.New("no model ids found in remote payload")
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
