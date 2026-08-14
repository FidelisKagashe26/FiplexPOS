package shops

import (
	"POS-fiplex/internal/common"
	"POS-fiplex/pkg/validator"
	"github.com/gofiber/fiber/v3"
)

type ShopHandler struct {
	service  IShopService
	validate validator.Validator
}

func NewShopHandler(service IShopService, validate validator.Validator) *ShopHandler {
	return &ShopHandler{
		service: service,
		validate: validate,
	}
}

// @Summary      Create Shop
// @Description  Create a new shop tenant (Super Admin only)
// @Tags         shops
// @Accept       json
// @Produce      json
// @Param        request body CreateShopRequest true "Shop creation data"
// @Success      201 {object} common.SuccessResponse{data=ShopResponse}
// @Failure      400 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /shops [post]
// @Security     BearerAuth
func (h *ShopHandler) CreateShop(c fiber.Ctx) error {
	var req CreateShopRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "invalid request body", Error: err.Error()})
	}

	if err := h.validate.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{Message: "validation failed", Error: err.Error()})
	}

	res, err := h.service.CreateShop(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to create shop"})
	}

	return c.Status(fiber.StatusCreated).JSON(common.SuccessResponse{
		Message: "Shop created successfully",
		Data:    res,
	})
}

// @Summary      List Shops
// @Description  List all shops (Super Admin only)
// @Tags         shops
// @Produce      json
// @Success      200 {object} common.SuccessResponse{data=ListShopsResponse}
// @Failure      500 {object} common.ErrorResponse
// @Router       /shops [get]
// @Security     BearerAuth
func (h *ShopHandler) ListShops(c fiber.Ctx) error {
	res, err := h.service.ListShops(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{Message: "failed to fetch shops"})
	}

	return c.Status(fiber.StatusOK).JSON(common.SuccessResponse{
		Data:    res.Data,
	})
}
