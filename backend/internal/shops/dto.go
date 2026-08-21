package shops

import (
	"github.com/google/uuid"
	"time"
)

type CreateShopRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=100"`
	Address   string `json:"address"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Phone     string `json:"phone" validate:"required,min=7,max=32"`
	OwnerName string `json:"owner_name" validate:"required,min=2,max=100"`
	Password  string `json:"password" validate:"required,min=8,max=100"`
}

type ShopResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	OwnerID   *uuid.UUID `json:"owner_id"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type ListShopsResponse struct {
	Data []ShopResponse `json:"data"`
}
