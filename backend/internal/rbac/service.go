package rbac

import (
	"context"
	"errors"

	"POS-fiplex/internal/authz"
	"POS-fiplex/internal/common"
	"POS-fiplex/internal/rbac/repository"
	"POS-fiplex/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrShopContext           = errors.New("an active shop is required")
	ErrRoleOutsideActiveShop = errors.New("role does not belong to the active shop")
)

type IRBACService interface {
	CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error)
	ListRoles(ctx context.Context) (*ListRolesResponse, error)
	ListPermissions(ctx context.Context) ([]PermissionResponse, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]PermissionResponse, error)
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	AssignUserRole(ctx context.Context, req AssignUserRoleRequest) error
}

type RBACService struct {
	repo     *repository.Queries
	pool     *pgxpool.Pool
	resolver *authz.Resolver
	log      logger.ILogger
}

func NewRBACService(repo *repository.Queries, pool *pgxpool.Pool, resolver *authz.Resolver, log logger.ILogger) IRBACService {
	return &RBACService{repo: repo, pool: pool, resolver: resolver, log: log}
}

func activeShopID(ctx context.Context) (uuid.UUID, error) {
	shopID, ok := common.ShopIDFromContext(ctx)
	if !ok {
		return uuid.Nil, ErrShopContext
	}
	return shopID, nil
}

// roleInActiveShop guarantees every role read, changed, or assigned belongs to
// the selected tenant. A role can never be shared by two shops.
func (s *RBACService) roleInActiveShop(ctx context.Context, roleID uuid.UUID) (uuid.UUID, error) {
	shopID, err := activeShopID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return uuid.Nil, err
	}
	if !role.ShopID.Valid || uuid.UUID(role.ShopID.Bytes) != shopID {
		return uuid.Nil, ErrRoleOutsideActiveShop
	}
	return shopID, nil
}

// CreateRole always attaches the role to the active shop. The legacy request
// shop_id is ignored, preventing a browser from creating global/foreign roles.
func (s *RBACService) CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	shopID, err := activeShopID(ctx)
	if err != nil {
		return nil, err
	}
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	role, err := s.repo.CreateRole(ctx, repository.CreateRoleParams{
		ShopID: pgtype.UUID{Bytes: shopID, Valid: true}, Name: req.Name, Description: desc,
	})
	if err != nil {
		s.log.Errorf("CreateRole | failed to create role: %v", err)
		return nil, err
	}
	return toRoleResponse(role), nil
}

func (s *RBACService) ListRoles(ctx context.Context) (*ListRolesResponse, error) {
	shopID, err := activeShopID(ctx)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.ListRolesByShop(ctx, pgtype.UUID{Bytes: shopID, Valid: true})
	if err != nil {
		s.log.Errorf("ListRoles | failed to list roles: %v", err)
		return nil, err
	}
	res := &ListRolesResponse{Data: make([]RoleResponse, 0, len(roles))}
	for _, role := range roles {
		if role.ShopID.Valid && uuid.UUID(role.ShopID.Bytes) == shopID {
			res.Data = append(res.Data, *toRoleResponse(role))
		}
	}
	return res, nil
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]PermissionResponse, error) {
	perms, err := s.repo.ListPermissions(ctx)
	if err != nil {
		s.log.Errorf("ListPermissions | failed to list permissions: %v", err)
		return nil, err
	}
	res := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		res = append(res, toPermissionResponse(p))
	}
	return res, nil
}

func (s *RBACService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]PermissionResponse, error) {
	if _, err := s.roleInActiveShop(ctx, roleID); err != nil {
		return nil, err
	}
	perms, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		s.log.Errorf("GetRolePermissions | failed to fetch role permissions: %v", err)
		return nil, err
	}
	res := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		res = append(res, toPermissionResponse(p))
	}
	return res, nil
}

// SetRolePermissions atomically replaces the permission matrix for one role in
// the active shop. It cannot alter a role from a different tenant.
func (s *RBACService) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if _, err := s.roleInActiveShop(ctx, roleID); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.repo.WithTx(tx)
	if err := qtx.ClearRolePermissions(ctx, roleID); err != nil {
		return err
	}
	for _, pid := range permissionIDs {
		if err := qtx.AssignPermissionToRole(ctx, repository.AssignPermissionToRoleParams{RoleID: roleID, PermissionID: pid}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AssignUserRole derives the tenant from trusted request context; client input
// cannot attach a staff account to a role in a different shop.
func (s *RBACService) AssignUserRole(ctx context.Context, req AssignUserRoleRequest) error {
	shopID, err := s.roleInActiveShop(ctx, req.RoleID)
	if err != nil {
		return err
	}
	if err := s.repo.AssignUserRole(ctx, repository.AssignUserRoleParams{UserID: req.UserID, ShopID: shopID, RoleID: req.RoleID}); err != nil {
		s.log.Errorf("AssignUserRole | failed: %v", err)
		return err
	}
	s.resolver.Invalidate(ctx, req.UserID, shopID)
	return nil
}

func toRoleResponse(role repository.Role) *RoleResponse {
	var shopID *uuid.UUID
	if role.ShopID.Valid {
		id := uuid.UUID(role.ShopID.Bytes)
		shopID = &id
	}
	var desc string
	if role.Description != nil {
		desc = *role.Description
	}
	return &RoleResponse{ID: role.ID, ShopID: shopID, Name: role.Name, Description: desc, CreatedAt: role.CreatedAt.Time}
}

func toPermissionResponse(p repository.Permission) PermissionResponse {
	var desc string
	if p.Description != nil {
		desc = *p.Description
	}
	return PermissionResponse{ID: p.ID, Name: p.Name, Module: p.Module, Description: desc}
}
