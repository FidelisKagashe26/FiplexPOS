package authz

import (
	"fmt"

	"POS-fiplex/internal/common"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Locals / header keys shared with AuthMiddleware (which sets user_id and role).
const (
	LocalUserID  = "user_id"
	LocalRole    = "role"
	LocalShopID  = "shop_id"
	HeaderShopID = "X-Shop-Id"

	// Only the platform role bypasses tenant permissions. An admin is the owner.
	// Shop owners must remain inside their shop's permission boundary.
	roleSuperAdmin = "superadmin"
)

func userIDFromCtx(c fiber.Ctx) (uuid.UUID, bool) {
	switch v := c.Locals(LocalUserID).(type) {
	case uuid.UUID:
		return v, v != uuid.Nil
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	default:
		return uuid.Nil, false
	}
}

// roleString reads the coarse role regardless of its concrete type
// (middleware.UserRole or string), keeping authz decoupled from the middleware pkg.
func roleString(c fiber.Ctx) string {
	v := c.Locals(LocalRole)
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func isPrivileged(c fiber.Ctx) bool {
	return roleString(c) == roleSuperAdmin
}

// stashShop makes the resolved shop available both to Fiber Locals and to the
// request context.Context, so downstream services can scope queries by shop_id.
func stashShop(c fiber.Ctx, shopID uuid.UUID) {
	c.Locals(LocalShopID, shopID)
	c.RequestCtx().SetUserValue(common.ShopIDKey, shopID)
}

// resolveShopID determines which shop the current request acts within:
// an already-resolved local, an explicit X-Shop-Id header, or the user's primary shop.
func (r *Resolver) resolveShopID(c fiber.Ctx, userID uuid.UUID) (uuid.UUID, bool) {
	if id, ok := c.Locals(LocalShopID).(uuid.UUID); ok && id != uuid.Nil {
		return id, true
	}
	if h := c.Get(HeaderShopID); h != "" {
		if id, err := uuid.Parse(h); err == nil {
			return id, true
		}
	}
	if id, err := r.PrimaryShop(c.Context(), userID); err == nil && id != uuid.Nil {
		return id, true
	}
	return uuid.Nil, false
}

// TenantResolver resolves the active shop for the request and stashes it in
// Locals for downstream handlers. Non-privileged users are rejected if they
// request (via X-Shop-Id) a shop they do not belong to. It is lenient when no
// shop can be determined (some routes are shop-agnostic); permission guards make
// the final call.
func (r *Resolver) TenantResolver() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := userIDFromCtx(c)
		if !ok {
			return c.Next() // AuthMiddleware is responsible for authentication.
		}

		var shopID uuid.UUID
		if h := c.Get(HeaderShopID); h != "" {
			id, err := uuid.Parse(h)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid X-Shop-Id"})
			}
			if !isPrivileged(c) {
				member, err := r.IsMember(c.Context(), userID, id)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to verify shop membership"})
				}
				if !member {
					return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{Message: "you have no access to this shop"})
				}
			}
			shopID = id
		}

		if shopID == uuid.Nil {
			if id, err := r.PrimaryShop(c.Context(), userID); err == nil {
				shopID = id
			}
		}

		if shopID != uuid.Nil {
			stashShop(c, shopID)
		}
		return c.Next()
	}
}

// RequirePermission enforces that the current user holds a permission. Admins
// (coarse owner tier) bypass the check. Enforcement is always server-side.
func (r *Resolver) RequirePermission(perm string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := userIDFromCtx(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{Message: "unauthorized"})
		}

		// Resolve the shop first (best-effort) so it is available to services for
		// data scoping even for admins, who target a shop via X-Shop-Id.
		shopID, hasShop := r.resolveShopID(c, userID)
		if hasShop {
			stashShop(c, shopID)
		}

		if isPrivileged(c) {
			return c.Next()
		}
		if !hasShop {
			return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{Message: "no shop context for this request"})
		}

		allowed, err := r.Has(c.Context(), userID, shopID, perm)
		if err != nil {
			r.log.Errorf("authz | permission check failed for user %s: %v", userID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "authorization check failed"})
		}
		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{Message: "you do not have permission to perform this action"})
		}
		return c.Next()
	}
}

// RequireAny enforces that the user holds at least one of the given permissions.
func (r *Resolver) RequireAny(perms ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := userIDFromCtx(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{Message: "unauthorized"})
		}

		shopID, hasShop := r.resolveShopID(c, userID)
		if hasShop {
			stashShop(c, shopID)
		}

		if isPrivileged(c) {
			return c.Next()
		}
		if !hasShop {
			return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{Message: "no shop context for this request"})
		}

		set, err := r.Effective(c.Context(), userID, shopID)
		if err != nil {
			r.log.Errorf("authz | permission check failed for user %s: %v", userID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "authorization check failed"})
		}
		for _, p := range perms {
			if _, ok := set[p]; ok {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{Message: "you do not have permission to perform this action"})
	}
}
