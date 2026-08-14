package shops

import (
	"POS-fiplex/internal/common"
	"POS-fiplex/internal/shops/repository"
	user_repo "POS-fiplex/internal/user/repository"
	"POS-fiplex/pkg/logger"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type IShopService interface {
	CreateShop(ctx context.Context, req CreateShopRequest) (*ShopResponse, error)
	ListShops(ctx context.Context) (*ListShopsResponse, error)
}

// OwnerLookup is the slice of the user repository the shops service needs to
// resolve an owner's email to their user id.
type OwnerLookup interface {
	GetUserByEmail(ctx context.Context, email string) (user_repo.User, error)
}

type ShopService struct {
	repo  repository.Querier
	users OwnerLookup
	log   logger.ILogger
}

func NewShopService(repo repository.Querier, users OwnerLookup, log logger.ILogger) IShopService {
	return &ShopService{
		repo:  repo,
		users: users,
		log:   log,
	}
}

func (s *ShopService) CreateShop(ctx context.Context, req CreateShopRequest) (*ShopResponse, error) {
	var address *string
	if req.Address != "" {
		address = &req.Address
	}

	// shops.owner_id is NOT NULL and references users(id); resolve the owner by email.
	owner, err := s.users.GetUserByEmail(ctx, req.OwnerEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		s.log.Errorf("CreateShop | failed to look up owner %q: %v", req.OwnerEmail, err)
		return nil, err
	}

	shop, err := s.repo.CreateShop(ctx, repository.CreateShopParams{
		Name:     req.Name,
		Address:  address,
		OwnerID:  owner.ID,
		IsActive: true,
	})
	if err != nil {
		s.log.Errorf("CreateShop | failed to create shop: %v", err)
		return nil, err
	}

	var resAddress string
	if shop.Address != nil {
		resAddress = *shop.Address
	}

	return &ShopResponse{
		ID:        shop.ID,
		Name:      shop.Name,
		Address:   resAddress,
		OwnerID:   &shop.OwnerID,
		IsActive:  shop.IsActive,
		CreatedAt: shop.CreatedAt.Time,
	}, nil
}

func (s *ShopService) ListShops(ctx context.Context) (*ListShopsResponse, error) {
	shops, err := s.repo.ListAllShops(ctx)
	if err != nil {
		s.log.Errorf("ListShops | failed to list shops: %v", err)
		return nil, err
	}

	res := &ListShopsResponse{
		Data: make([]ShopResponse, 0, len(shops)),
	}

	for _, shop := range shops {
		var resAddress string
		if shop.Address != nil {
			resAddress = *shop.Address
		}
		
		owner := shop.OwnerID

		res.Data = append(res.Data, ShopResponse{
			ID:        shop.ID,
			Name:      shop.Name,
			Address:   resAddress,
			OwnerID:   &owner,
			IsActive:  shop.IsActive,
			CreatedAt: shop.CreatedAt.Time,
		})
	}

	return res, nil
}
