package common

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	// ShopIDKey carries the active shop for the request, set by the tenant/permission
	// middleware and read by services to scope queries by shop_id (data isolation).
	ShopIDKey contextKey = "shopID"
)

// ShopIDFromContext returns the active shop id for the request, if one was
// resolved. Services use it to isolate data per tenant.
func ShopIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ShopIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// ShopParamFromContext returns the active shop as a pgtype.UUID for use as a
// query parameter. When no shop is active it returns an invalid (SQL NULL) value,
// which the lenient "shop_id = $ OR $ IS NULL" query pattern treats as "all shops"
// — preserving single-tenant behaviour until shops are in use.
func ShopParamFromContext(ctx context.Context) pgtype.UUID {
	if id, ok := ShopIDFromContext(ctx); ok {
		return pgtype.UUID{Bytes: id, Valid: true}
	}
	return pgtype.UUID{Valid: false}
}
