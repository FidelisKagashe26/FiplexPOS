package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rbac_repo "POS-fiplex/internal/rbac/repository"
	"POS-fiplex/pkg/cache"
	"POS-fiplex/pkg/logger"

	"github.com/google/uuid"
)

// permCacheTTL is deliberately short. A user's effective permissions are cached
// per (user, shop) to avoid a DB round-trip on every request, but role/permission
// edits should take effect quickly, so we cap staleness at this window (in
// addition to the explicit Invalidate on user-role assignment).
const permCacheTTL = 60 * time.Second

// Resolver resolves and caches a user's effective permissions within a shop.
// It is the backend authority for every authorization decision.
type Resolver struct {
	repo  *rbac_repo.Queries
	cache cache.Cache
	log   logger.ILogger
}

// NewResolver builds a Resolver over the RBAC repository and Redis cache.
func NewResolver(repo *rbac_repo.Queries, c cache.Cache, log logger.ILogger) *Resolver {
	return &Resolver{repo: repo, cache: c, log: log}
}

func permCacheKey(userID, shopID uuid.UUID) string {
	return fmt.Sprintf("authz:perms:%s:%s", userID.String(), shopID.String())
}

// Effective returns the set of permission names the user holds in the shop.
// The result is cached in Redis for permCacheTTL. A nil/empty set means the user
// has no fine-grained permissions in that shop.
func (r *Resolver) Effective(ctx context.Context, userID, shopID uuid.UUID) (map[string]struct{}, error) {
	key := permCacheKey(userID, shopID)

	if raw, err := r.cache.GetWithContext(ctx, key); err == nil && len(raw) > 0 {
		var names []string
		if jsonErr := json.Unmarshal(raw, &names); jsonErr == nil {
			return toSet(names), nil
		}
		// Corrupt cache entry: fall through and refetch.
	}

	names, err := r.repo.GetUserPermissionNamesInShop(ctx, userID, shopID)
	if err != nil {
		return nil, err
	}

	if raw, jsonErr := json.Marshal(names); jsonErr == nil {
		if setErr := r.cache.SetWithContext(ctx, key, raw, permCacheTTL); setErr != nil {
			r.log.Warnf("authz | failed to cache permissions for user %s shop %s: %v", userID, shopID, setErr)
		}
	}

	return toSet(names), nil
}

// Has reports whether the user holds a specific permission in the shop.
func (r *Resolver) Has(ctx context.Context, userID, shopID uuid.UUID, perm string) (bool, error) {
	set, err := r.Effective(ctx, userID, shopID)
	if err != nil {
		return false, err
	}
	_, ok := set[perm]
	return ok, nil
}

// Invalidate clears the cached permission set for a user in a shop. Call this
// whenever the user's role assignment changes.
func (r *Resolver) Invalidate(ctx context.Context, userID, shopID uuid.UUID) {
	if err := r.cache.DeleteWithContext(ctx, permCacheKey(userID, shopID)); err != nil {
		r.log.Warnf("authz | failed to invalidate permission cache for user %s shop %s: %v", userID, shopID, err)
	}
}

// PrimaryShop returns the shop a user belongs to when none is specified explicitly.
func (r *Resolver) PrimaryShop(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.repo.GetUserPrimaryShopID(ctx, userID)
}

// IsMember reports whether the user has a role assignment in the shop.
func (r *Resolver) IsMember(ctx context.Context, userID, shopID uuid.UUID) (bool, error) {
	return r.repo.IsUserInShop(ctx, userID, shopID)
}

// SeedPermissions upserts every permission in the catalog. It is idempotent and
// safe to run on every startup; the Go catalog stays the source of truth.
func (r *Resolver) SeedPermissions(ctx context.Context) error {
	for _, entry := range Catalog {
		desc := entry.Description
		var descPtr *string
		if desc != "" {
			descPtr = &desc
		}
		if _, err := r.repo.UpsertPermission(ctx, entry.Name, entry.Module, descPtr); err != nil {
			return fmt.Errorf("seed permission %q: %w", entry.Name, err)
		}
	}
	r.log.Infof("authz | seeded %d permissions from catalog", len(Catalog))
	return nil
}

func toSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return set
}
