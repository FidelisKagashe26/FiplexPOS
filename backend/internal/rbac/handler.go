package rbac

import (
	"fmt"

	"POS-fiplex/internal/authz"
	"POS-fiplex/internal/common"
	"POS-fiplex/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type RBACHandler struct {
	service  IRBACService
	resolver *authz.Resolver
	validate validator.Validator
}

func NewRBACHandler(service IRBACService, resolver *authz.Resolver, validate validator.Validator) *RBACHandler {
	return &RBACHandler{
		service:  service,
		resolver: resolver,
		validate: validate,
	}
}

// CreateRole creates a new role for a shop.
// @Router /roles [post]
func (h *RBACHandler) CreateRole(c fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid request body", Error: err.Error()})
	}
	if err := h.validate.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "validation failed", Error: err.Error()})
	}

	res, err := h.service.CreateRole(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to create role"})
	}
	return c.Status(fiber.StatusCreated).JSON(common.SuccessResponse{Message: "Role created successfully", Data: res})
}

// ListRoles lists all roles.
// @Router /roles [get]
func (h *RBACHandler) ListRoles(c fiber.Ctx) error {
	res, err := h.service.ListRoles(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to fetch roles"})
	}
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: res.Data})
}

// ListPermissions returns the full permission catalog for the frontend matrix.
// @Router /permissions [get]
func (h *RBACHandler) ListPermissions(c fiber.Ctx) error {
	res, err := h.service.ListPermissions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to fetch permissions"})
	}
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: res})
}

// GetRolePermissions returns the permissions currently attached to a role.
// @Router /roles/{id}/permissions [get]
func (h *RBACHandler) GetRolePermissions(c fiber.Ctx) error {
	roleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid role id"})
	}
	res, err := h.service.GetRolePermissions(c.Context(), roleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to fetch role permissions"})
	}
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: res})
}

// SetRolePermissions replaces a role's permission matrix. This is the endpoint
// the frontend "Save Permissions" button calls.
// @Router /roles/{id}/permissions [put]
func (h *RBACHandler) SetRolePermissions(c fiber.Ctx) error {
	roleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid role id"})
	}
	var req UpdateRolePermissionsRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid request body", Error: err.Error()})
	}
	if err := h.validate.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "validation failed", Error: err.Error()})
	}
	if err := h.service.SetRolePermissions(c.Context(), roleID, req.PermissionIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to update role permissions"})
	}
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Message: "Permissions updated successfully"})
}

// AssignUserRole links a user to a role within a shop.
// @Router /roles/assign [post]
func (h *RBACHandler) AssignUserRole(c fiber.Ctx) error {
	var req AssignUserRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid request body", Error: err.Error()})
	}
	if err := h.validate.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "validation failed", Error: err.Error()})
	}
	if err := h.service.AssignUserRole(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to assign role"})
	}
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Message: "Role assigned successfully"})
}

// MyPermissions returns the effective permissions of the current user, for the
// frontend to render (show/hide) UI. Admins receive the full catalog. The
// backend still enforces every action independently.
// @Router /auth/me/permissions [get]
func (h *RBACHandler) MyPermissions(c fiber.Ctx) error {
	userVal := c.Locals(authz.LocalUserID)
	var userID uuid.UUID
	switch v := userVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{Message: "unauthorized"})
		}
		userID = id
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{Message: "unauthorized"})
	}

	role := fmt.Sprintf("%v", c.Locals(authz.LocalRole))

	// Admin (owner tier) implicitly holds every permission.
	if role == "admin" {
		return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: MyPermissionsResponse{
			Role:        role,
			Permissions: authz.AllPermissionNames(),
		}})
	}

	shopID, err := h.resolver.PrimaryShop(c.Context(), userID)
	if err != nil {
		// No shop assignment yet → no fine-grained permissions.
		return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: MyPermissionsResponse{
			Role:        role,
			Permissions: []string{},
		}})
	}

	set, err := h.resolver.Effective(c.Context(), userID, shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to resolve permissions"})
	}
	perms := make([]string, 0, len(set))
	for p := range set {
		perms = append(perms, p)
	}
	sid := shopID
	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{Data: MyPermissionsResponse{
		Role:        role,
		ShopID:      &sid,
		Permissions: perms,
	}})
}
