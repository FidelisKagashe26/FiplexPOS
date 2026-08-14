package shops

import (
	"time"
	"github.com/google/uuid"
)

type CreateShopRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Address     string `json:"address"`
	OwnerEmail  string `json:"owner_email" validate:"required,email"`
}

type ShopResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	OwnerID   *uuid.UUID `json:"owner_id"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type ListShopsResponse struct {
	Data []ShopResponse `json:"data"`
}
