package server

import (
	"POS-fiplex/internal/authz"
	"POS-fiplex/internal/common/middleware"
	ws "POS-fiplex/internal/websocket"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *App, container *AppContainer) {
	hltHandler := HealthHandler(app)
	app.FiberApp.Get("/healthz", hltHandler)

	api := app.FiberApp.Group("/api/v1")

	api.Use(middleware.RateLimiter(app.RedisCache))

	api.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	authMiddleware := middleware.AuthMiddleware(app.JWT, app.Logger)
	idempotencyMiddleware := middleware.Idempotency(app.RedisCache)

	// Authorization guards are enforced entirely on the backend by the authz
	// resolver. RequirePermission(<perm>) checks the user's effective permissions
	// for their shop; admins (coarse owner tier) bypass. The frontend only renders
	// from the permission list it receives — it never decides access.
	rbacGuard := container.Authz

	api.Get("/ws", authMiddleware, websocket.New(func(c *websocket.Conn) {
		client := ws.NewClient(container.WSHub, c)
		container.WSHub.Register(client)

		go client.WritePump()
		client.ReadPump()
	}))

	api.Post("/auth/login", container.AuthHandler.LoginHandler)
	api.Post("/auth/refresh", container.AuthHandler.RefreshHandler)
	api.Get("/auth/me", authMiddleware, container.AuthHandler.ProfileHandler)
	api.Get("/auth/me/permissions", authMiddleware, container.RBACHandler.MyPermissions)
	api.Post("/auth/add", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersCreate), container.AuthHandler.AddUserHandler)
	api.Put("/auth/me/avatar", authMiddleware, container.AuthHandler.UpdateAvatarHandler)
	api.Put("/auth/me/password", authMiddleware, container.AuthHandler.UpdatePasswordHandler)
	api.Post("/auth/logout", authMiddleware, container.AuthHandler.LogoutHandler)

	api.Get("/users", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersView), container.UserHandler.GetAllUsersHandler)
	api.Post("/users", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersCreate), container.UserHandler.CreateUserHandler)
	api.Get("/users/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersView), container.UserHandler.GetUserByIDHandler)
	api.Put("/users/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersEdit), container.UserHandler.UpdateUserHandler)
	api.Post("/users/:id/toggle-status", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersEdit), container.UserHandler.ToggleUserStatusHandler)
	api.Delete("/users/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermUsersDelete), container.UserHandler.DeleteUserHandler)

	api.Get("/categories", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesView), container.CategoryHandler.GetAllCategoriesHandler)
	api.Get("/categories/count", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesView), container.CategoryHandler.GetCategoryCountHandler)
	api.Post("/categories", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesCreate), container.CategoryHandler.CreateCategoryHandler)
	api.Get("/categories/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesView), container.CategoryHandler.GetCategoryByIDHandler)
	api.Put("/categories/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesEdit), container.CategoryHandler.UpdateCategoryHandler)
	api.Delete("/categories/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermCategoriesDelete), container.CategoryHandler.DeleteCategoryHandler)

	api.Post("/products", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsCreate), container.ProductHandler.CreateProductHandler)
	api.Post("/products/:id/image", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.UploadProductImageHandler)

	// Deleted Products Management
	api.Get("/products/trash", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsRestore), container.ProductHandler.ListDeletedProductsHandler)
	api.Post("/products/trash/restore-bulk", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsRestore), container.ProductHandler.RestoreProductsBulkHandler)
	api.Get("/products/trash/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsRestore), container.ProductHandler.GetDeletedProductHandler)
	api.Post("/products/trash/:id/restore", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsRestore), container.ProductHandler.RestoreProductHandler)

	api.Get("/products", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsView), container.ProductHandler.ListProductsHandler)
	api.Get("/products/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsView), container.ProductHandler.GetProductHandler)
	api.Get("/products/:id/stock-history", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsView), container.ProductHandler.GetStockHistoryHandler)
	api.Patch("/products/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.UpdateProductHandler)
	api.Delete("/products/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsDelete), container.ProductHandler.DeleteProductHandler)

	api.Post("/products/:product_id/options", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.CreateProductOptionHandler)
	api.Post("/products/:product_id/options/:option_id/image", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.UploadProductOptionImageHandler)
	api.Patch("/products/:product_id/options/:option_id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.UpdateProductOptionHandler)
	api.Delete("/products/:product_id/options/:option_id", authMiddleware, rbacGuard.RequirePermission(authz.PermProductsEdit), container.ProductHandler.DeleteProductOptionHandler)

	api.Get("/payment-methods", authMiddleware, container.PaymentMethodHandler.ListPaymentMethodsHandler)
	api.Get("/cancellation-reasons", authMiddleware, container.CancellationReasonHandler.ListCancellationReasonsHandler)
	api.Get("/activity-logs", authMiddleware, rbacGuard.RequirePermission(authz.PermActivityLogsView), container.ActivityLogHandler.GetActivityLogs)

	api.Post("/orders", authMiddleware, middleware.RequireIdempotencyKey(), idempotencyMiddleware, rbacGuard.RequirePermission(authz.PermOrdersCreate), middleware.ShiftMiddleware(container.ShiftRepo, app.Cache, app.Logger), container.OrderHandler.CreateOrderHandler)
	api.Get("/orders", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersView), container.OrderHandler.ListOrdersHandler)
	api.Get("/orders/:id", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersView), container.OrderHandler.GetOrderHandler)
	api.Patch("/orders/:id/items", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersEdit), container.OrderHandler.UpdateOrderItemsHandler)

	api.Post("/orders/:id/cancel", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersCancel), container.OrderHandler.CancelOrderHandler)
	api.Post("/orders/:id/refund", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersRefund), container.OrderHandler.RefundOrderHandler)
	api.Post("/orders/:id/apply-promotion", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersEdit), container.OrderHandler.ApplyPromotionHandler)
	api.Post("/orders/:id/pay/midtrans", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersPay), container.OrderHandler.InitiateMidtransPaymentHandler)
	api.Post("/orders/:id/pay/manual", authMiddleware, middleware.RequireIdempotencyKey(), idempotencyMiddleware, rbacGuard.RequirePermission(authz.PermOrdersPay), container.OrderHandler.ConfirmManualPaymentHandler)
	api.Post("/orders/:id/update-status", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersEdit), container.OrderHandler.UpdateOperationalStatusHandler)

	api.Post("/orders/:id/print", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersPrint), container.PrinterHandler.PrintInvoiceHandler)
	api.Get("/orders/:id/print-data", authMiddleware, rbacGuard.RequirePermission(authz.PermOrdersView), container.PrinterHandler.GetInvoiceDataHandler)
	api.Post("/payments/midtrans-notification", container.OrderHandler.MidtransNotificationHandler)

	reportsGroup := api.Group("/reports", authMiddleware, rbacGuard.TenantResolver(), rbacGuard.RequirePermission(authz.PermReportsView))
	{
		reportsGroup.Get("/dashboard-summary", container.ReportHandler.GetDashboardSummaryHandler)
		reportsGroup.Get("/sales", container.ReportHandler.GetSalesReportsHandler)
		reportsGroup.Get("/products", container.ReportHandler.GetProductPerformanceHandler)
		reportsGroup.Get("/payment-methods", container.ReportHandler.GetPaymentMethodPerformanceHandler)
		reportsGroup.Get("/cashier-performance", container.ReportHandler.GetCashierPerformanceHandler)
		reportsGroup.Get("/cancellations", container.ReportHandler.GetCancellationReportsHandler)
		reportsGroup.Get("/profit-summary", container.ReportHandler.GetProfitSummaryHandler)
		reportsGroup.Get("/profit-products", container.ReportHandler.GetProductProfitReportsHandler)
		reportsGroup.Get("/low-stock", container.ReportHandler.GetLowStockProductsHandler)
		reportsGroup.Get("/promotions", container.ReportHandler.GetPromotionPerformanceHandler)
		reportsGroup.Get("/shift-summary", container.ReportHandler.GetShiftSummaryHandler)
	}

	promotionsReadGroup := api.Group("/promotions", authMiddleware, rbacGuard.RequirePermission(authz.PermPromotionsView))
	{
		promotionsReadGroup.Get("/", container.PromotionHandler.ListPromotionsHandler)
		promotionsReadGroup.Get("/:id", container.PromotionHandler.GetPromotionHandler)
	}

	promotionsWriteGroup := api.Group("/promotions", authMiddleware)
	{
		promotionsWriteGroup.Post("/", rbacGuard.RequirePermission(authz.PermPromotionsCreate), container.PromotionHandler.CreatePromotionHandler)
		promotionsWriteGroup.Put("/:id", rbacGuard.RequirePermission(authz.PermPromotionsEdit), container.PromotionHandler.UpdatePromotionHandler)
		promotionsWriteGroup.Delete("/:id", rbacGuard.RequirePermission(authz.PermPromotionsDelete), container.PromotionHandler.DeletePromotionHandler)
		promotionsWriteGroup.Post("/:id/restore", rbacGuard.RequirePermission(authz.PermPromotionsEdit), container.PromotionHandler.RestorePromotionHandler)
	}

	settingsGroup := api.Group("/settings", authMiddleware)
	{
		settingsGroup.Get("/branding", rbacGuard.RequirePermission(authz.PermSettingsView), container.SettingsHandler.GetBrandingHandler)
		settingsGroup.Put("/branding", rbacGuard.RequirePermission(authz.PermSettingsEdit), container.SettingsHandler.UpdateBrandingHandler)
		settingsGroup.Post("/branding/logo", rbacGuard.RequirePermission(authz.PermSettingsEdit), container.SettingsHandler.UpdateLogoHandler)

		settingsGroup.Get("/printer", rbacGuard.RequirePermission(authz.PermSettingsView), container.SettingsHandler.GetPrinterSettingsHandler)
		settingsGroup.Put("/printer", rbacGuard.RequirePermission(authz.PermSettingsEdit), container.SettingsHandler.UpdatePrinterSettingsHandler)
		settingsGroup.Get("/printer/discover", rbacGuard.RequirePermission(authz.PermSettingsEdit), container.PrinterHandler.DiscoverPrintersHandler)
		settingsGroup.Post("/printer/test", rbacGuard.RequirePermission(authz.PermSettingsEdit), container.PrinterHandler.TestPrintHandler)
	}

	shiftGroup := api.Group("/shifts", authMiddleware, rbacGuard.RequirePermission(authz.PermShiftsManage))
	{
		shiftGroup.Post("/start", container.ShiftHandler.StartShiftHandler)
		shiftGroup.Post("/end", container.ShiftHandler.EndShiftHandler)
		shiftGroup.Get("/current", container.ShiftHandler.GetOpenShiftHandler)
		shiftGroup.Post("/cash-transaction", container.ShiftHandler.CreateCashTransactionHandler)
	}

	customerGroup := api.Group("/customers", authMiddleware)
	{
		customerGroup.Get("/", rbacGuard.RequirePermission(authz.PermCustomersView), container.CustomerHandler.ListCustomersHandler)
		customerGroup.Post("/", rbacGuard.RequirePermission(authz.PermCustomersCreate), container.CustomerHandler.CreateCustomerHandler)
		customerGroup.Get("/:id", rbacGuard.RequirePermission(authz.PermCustomersView), container.CustomerHandler.GetCustomerHandler)
		customerGroup.Put("/:id", rbacGuard.RequirePermission(authz.PermCustomersEdit), container.CustomerHandler.UpdateCustomerHandler)
		customerGroup.Delete("/:id", rbacGuard.RequirePermission(authz.PermCustomersDelete), container.CustomerHandler.DeleteCustomerHandler)
	}

	// Shop administration is a super-admin / owner concern. TenantResolver lets an
	// admin target a specific shop via X-Shop-Id.
	shopsGroup := api.Group("/shops", authMiddleware, rbacGuard.TenantResolver())
	{
		shopsGroup.Post("/", rbacGuard.RequirePermission(authz.PermShopsManage), container.ShopHandler.CreateShop)
		shopsGroup.Get("/", rbacGuard.RequirePermission(authz.PermShopsView), container.ShopHandler.ListShops)
	}

	rolesGroup := api.Group("/roles", authMiddleware, rbacGuard.TenantResolver())
	{
		rolesGroup.Post("/", rbacGuard.RequirePermission(authz.PermRolesManage), container.RBACHandler.CreateRole)
		rolesGroup.Get("/", rbacGuard.RequirePermission(authz.PermRolesView), container.RBACHandler.ListRoles)
		rolesGroup.Post("/assign", rbacGuard.RequirePermission(authz.PermRolesManage), container.RBACHandler.AssignUserRole)
		rolesGroup.Get("/:id/permissions", rbacGuard.RequirePermission(authz.PermRolesView), container.RBACHandler.GetRolePermissions)
		rolesGroup.Put("/:id/permissions", rbacGuard.RequirePermission(authz.PermRolesManage), container.RBACHandler.SetRolePermissions)
	}

	api.Get("/permissions", authMiddleware, rbacGuard.RequirePermission(authz.PermRolesView), container.RBACHandler.ListPermissions)

	SetupFrontend(app)
}
