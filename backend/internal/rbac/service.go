package rbac

import (
	"context"

	"POS-fiplex/internal/authz"
	"POS-fiplex/internal/rbac/repository"
	"POS-fiplex/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	return &RBACService{
		repo:     repo,
		pool:     pool,
		resolver: resolver,
		log:      log,
	}
}

func (s *RBACService) CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	var shopID pgtype.UUID
	if req.ShopID != nil {
		shopID = pgtype.UUID{Bytes: *req.ShopID, Valid: true}
	}

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	role, err := s.repo.CreateRole(ctx, repository.CreateRoleParams{
		ShopID:      shopID,
		Name:        req.Name,
		Description: desc,
	})
	if err != nil {
		s.log.Errorf("CreateRole | failed to create role: %v", err)
		return nil, err
	}

	return toRoleResponse(role), nil
}

func (s *RBACService) ListRoles(ctx context.Context) (*ListRolesResponse, error) {
	roles, err := s.repo.ListRolesByShop(ctx, pgtype.UUID{Valid: false})
	if err != nil {
		s.log.Errorf("ListRoles | failed to list roles: %v", err)
		return nil, err
	}

	res := &ListRolesResponse{Data: make([]RoleResponse, 0, len(roles))}
	for _, role := range roles {
		res.Data = append(res.Data, *toRoleResponse(role))
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

// SetRolePermissions atomically replaces a role's permission set with the given
// permission ids. Runs in a transaction so a partial update can never leave the
// matrix half-applied.
func (s *RBACService) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Errorf("SetRolePermissions | begin tx: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	if err := qtx.ClearRolePermissions(ctx, roleID); err != nil {
		s.log.Errorf("SetRolePermissions | clear: %v", err)
		return err
	}

	for _, pid := range permissionIDs {
		if err := qtx.AssignPermissionToRole(ctx, repository.AssignPermissionToRoleParams{
			RoleID:       roleID,
			PermissionID: pid,
		}); err != nil {
			s.log.Errorf("SetRolePermissions | assign %s: %v", pid, err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Errorf("SetRolePermissions | commit: %v", err)
		return err
	}
	return nil
}

func (s *RBACService) AssignUserRole(ctx context.Context, req AssignUserRoleRequest) error {
	if err := s.repo.AssignUserRole(ctx, repository.AssignUserRoleParams{
		UserID: req.UserID,
		ShopID: req.ShopID,
		RoleID: req.RoleID,
	}); err != nil {
		s.log.Errorf("AssignUserRole | failed: %v", err)
		return err
	}
	// The user's cached permission set for this shop is now stale.
	s.resolver.Invalidate(ctx, req.UserID, req.ShopID)
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
	return &RoleResponse{
		ID:          role.ID,
		ShopID:      shopID,
		Name:        role.Name,
		Description: desc,
		CreatedAt:   role.CreatedAt.Time,
	}
}

func toPermissionResponse(p repository.Permission) PermissionResponse {
	var desc string
	if p.Description != nil {
		desc = *p.Description
	}
	return PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Module:      p.Module,
		Description: desc,
	}
}
