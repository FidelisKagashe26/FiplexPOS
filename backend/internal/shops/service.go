package shops

import (
	"POS-fiplex/internal/shops/repository"
	"POS-fiplex/pkg/logger"
	"POS-fiplex/pkg/utils"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IShopService interface {
	CreateShop(ctx context.Context, req CreateShopRequest) (*ShopResponse, error)
	ListShops(ctx context.Context) (*ListShopsResponse, error)
}

type ShopService struct {
	repo repository.Querier
	pool *pgxpool.Pool
	log  logger.ILogger
}

func NewShopService(repo repository.Querier, pool *pgxpool.Pool, log logger.ILogger) IShopService {
	return &ShopService{repo: repo, pool: pool, log: log}
}

// CreateShop onboards a tenant and its owner in one atomic operation. The shop
// owns its email and phone; the owner uses the same email solely to log in.
func (s *ShopService) CreateShop(ctx context.Context, req CreateShopRequest) (*ShopResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ownerID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s-%s", strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.OwnerName), " ", "-")), ownerID.String()[:8])
	if _, err = tx.Exec(ctx, `INSERT INTO users (id, username, email, password_hash, role, is_active) VALUES ($1,$2,$3,$4,'admin',true)`, ownerID, username, req.Email, passwordHash); err != nil {
		return nil, err
	}

	res := &ShopResponse{OwnerID: &ownerID, IsActive: true}
	err = tx.QueryRow(ctx, `INSERT INTO shops (name,address,email,phone,owner_id,is_active) VALUES ($1,NULLIF($2,''),$3,$4,$5,true) RETURNING id,name,COALESCE(address,''),email,phone,created_at`, req.Name, req.Address, req.Email, req.Phone, ownerID).Scan(&res.ID, &res.Name, &res.Address, &res.Email, &res.Phone, &res.CreatedAt)
	if err != nil {
		return nil, err
	}

	var ownerRoleID uuid.UUID
	if err = tx.QueryRow(ctx, `INSERT INTO roles (shop_id,name,description) VALUES ($1,'Owner','Full access within this shop') RETURNING id`, res.ID).Scan(&ownerRoleID); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) SELECT $1,id FROM permissions WHERE module <> 'shops'`, ownerRoleID); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO user_shop_roles (user_id,shop_id,role_id) VALUES ($1,$2,$3)`, ownerID, res.ID, ownerRoleID); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ShopService) ListShops(ctx context.Context) (*ListShopsResponse, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,name,COALESCE(address,''),COALESCE(email,''),COALESCE(phone,''),owner_id,is_active,created_at FROM shops ORDER BY created_at DESC`)
	if err != nil {
		s.log.Errorf("ListShops | failed to list shops: %v", err)
		return nil, err
	}
	defer rows.Close()
	res := &ListShopsResponse{Data: []ShopResponse{}}
	for rows.Next() {
		var shop ShopResponse
		var owner uuid.UUID
		if err := rows.Scan(&shop.ID, &shop.Name, &shop.Address, &shop.Email, &shop.Phone, &owner, &shop.IsActive, &shop.CreatedAt); err != nil {
			return nil, err
		}
		shop.OwnerID = &owner
		res.Data = append(res.Data, shop)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
