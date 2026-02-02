package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jaxxstorm/landlord/internal/tenant"
	"go.uber.org/zap"
)

// LandlordClient fetches tenant data from the landlord API.
type LandlordClient interface {
	GetTenant(ctx context.Context, tenantUUID string) (*LandlordTenant, error)
}

// LandlordTenant represents the fields needed for compute resolution.
type LandlordTenant struct {
	Name          string                 `json:"name"`
	DesiredConfig map[string]interface{} `json:"desired_config,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Annotations   map[string]string      `json:"annotations,omitempty"`
}

type computeProviderCacheEntry struct {
	provider  string
	expiresAt time.Time
}

// CachedComputeProviderResolver resolves compute provider names with caching.
type CachedComputeProviderResolver struct {
	client   LandlordClient
	repo     tenant.Repository
	override string
	ttl      time.Duration
	logger   *zap.Logger

	mu    sync.RWMutex
	cache map[string]computeProviderCacheEntry
}

// NewCachedComputeProviderResolver creates a resolver with caching and optional override.
func NewCachedComputeProviderResolver(client LandlordClient, repo tenant.Repository, override string, ttl time.Duration, logger *zap.Logger) *CachedComputeProviderResolver {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &CachedComputeProviderResolver{
		client:   client,
		repo:     repo,
		override: override,
		ttl:      ttl,
		logger:   logger.With(zap.String("component", "compute-provider-resolver")),
		cache:    make(map[string]computeProviderCacheEntry),
	}
}

// ResolveProvider resolves the compute provider name for a tenant.
func (r *CachedComputeProviderResolver) ResolveProvider(ctx context.Context, tenantID, tenantUUID string) (string, error) {
	if r.override != "" {
		return r.override, nil
	}

	cacheKey := tenantUUID
	if cacheKey == "" {
		cacheKey = tenantID
	}

	if cacheKey != "" {
		if provider, ok := r.cachedProvider(cacheKey); ok {
			return provider, nil
		}
	}

	var provider string
	var err error

	if r.client != nil && tenantUUID != "" {
		provider, err = r.resolveFromAPI(ctx, tenantUUID)
		if err != nil {
			return "", err
		}
	} else if r.repo != nil && tenantID != "" {
		provider, err = r.resolveFromRepo(ctx, tenantID)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("no tenant lookup available for compute provider resolution")
	}

	if provider != "" && cacheKey != "" {
		r.setCachedProvider(cacheKey, provider)
	}

	return provider, nil
}

func (r *CachedComputeProviderResolver) cachedProvider(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.cache[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.provider, true
}

func (r *CachedComputeProviderResolver) setCachedProvider(key, provider string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[key] = computeProviderCacheEntry{
		provider:  provider,
		expiresAt: time.Now().Add(r.ttl),
	}
}

func (r *CachedComputeProviderResolver) resolveFromRepo(ctx context.Context, tenantName string) (string, error) {
	t, err := r.repo.GetTenantByName(ctx, tenantName)
	if err != nil {
		return "", fmt.Errorf("fetch tenant from repo: %w", err)
	}
	return providerFromMaps(t.DesiredConfig, t.Labels, t.Annotations), nil
}

func (r *CachedComputeProviderResolver) resolveFromAPI(ctx context.Context, tenantUUID string) (string, error) {
	tenantInfo, err := r.client.GetTenant(ctx, tenantUUID)
	if err != nil {
		return "", fmt.Errorf("fetch tenant from api: %w", err)
	}
	return providerFromMaps(tenantInfo.DesiredConfig, tenantInfo.Labels, tenantInfo.Annotations), nil
}

func providerFromMaps(config map[string]interface{}, labels map[string]string, annotations map[string]string) string {
	if config != nil {
		if provider, ok := config["compute_provider"]; ok {
			if value, ok := provider.(string); ok {
				return value
			}
		}
		if provider, ok := config["compute_provider_type"]; ok {
			if value, ok := provider.(string); ok {
				return value
			}
		}
	}
	if labels != nil {
		if provider, ok := labels["compute_provider"]; ok {
			return provider
		}
	}
	if annotations != nil {
		if provider, ok := annotations["compute_provider"]; ok {
			return provider
		}
	}
	return ""
}
