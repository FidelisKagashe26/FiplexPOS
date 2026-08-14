// Package authz is the single source of truth for role-based access control.
//
// All authorization is enforced on the backend. The frontend only receives the
// user's effective permission list (via /auth/me) and uses it to show/hide UI —
// it never decides access itself. Permissions are flat string constants grouped
// by module, stored in the permissions table and attached to roles; a user's
// effective set is resolved per request (and cached) from
// user_shop_roles -> roles -> role_permissions -> permissions.
package authz

// Permission is a flat, stable identifier of the form "module.action".
// Values are persisted in the permissions table (by name) and must not change
// once shipped, or previously-granted permissions would silently detach.
type Permission = string

const (
	// Products
	PermProductsView    Permission = "products.view"
	PermProductsCreate  Permission = "products.create"
	PermProductsEdit    Permission = "products.edit"
	PermProductsDelete  Permission = "products.delete"
	PermProductsRestore Permission = "products.restore"

	// Categories
	PermCategoriesView   Permission = "categories.view"
	PermCategoriesCreate Permission = "categories.create"
	PermCategoriesEdit   Permission = "categories.edit"
	PermCategoriesDelete Permission = "categories.delete"

	// Orders
	PermOrdersView   Permission = "orders.view"
	PermOrdersCreate Permission = "orders.create"
	PermOrdersEdit   Permission = "orders.edit"
	PermOrdersCancel Permission = "orders.cancel"
	PermOrdersRefund Permission = "orders.refund"
	PermOrdersPay    Permission = "orders.pay"
	PermOrdersPrint  Permission = "orders.print"

	// Customers
	PermCustomersView   Permission = "customers.view"
	PermCustomersCreate Permission = "customers.create"
	PermCustomersEdit   Permission = "customers.edit"
	PermCustomersDelete Permission = "customers.delete"

	// Promotions
	PermPromotionsView   Permission = "promotions.view"
	PermPromotionsCreate Permission = "promotions.create"
	PermPromotionsEdit   Permission = "promotions.edit"
	PermPromotionsDelete Permission = "promotions.delete"

	// Reports
	PermReportsView Permission = "reports.view"

	// Shifts
	PermShiftsManage Permission = "shifts.manage"

	// Users (staff of a shop)
	PermUsersView   Permission = "users.view"
	PermUsersCreate Permission = "users.create"
	PermUsersEdit   Permission = "users.edit"
	PermUsersDelete Permission = "users.delete"

	// Roles & permissions administration
	PermRolesView   Permission = "roles.view"
	PermRolesManage Permission = "roles.manage"

	// Shops (tenant administration)
	PermShopsView   Permission = "shops.view"
	PermShopsManage Permission = "shops.manage"

	// Settings
	PermSettingsView Permission = "settings.view"
	PermSettingsEdit Permission = "settings.edit"

	// Activity logs
	PermActivityLogsView Permission = "activity_logs.view"
)

// CatalogEntry describes a permission for seeding and for the frontend matrix.
type CatalogEntry struct {
	Name        Permission
	Module      string
	Description string
}

// Catalog is the complete, ordered list of permissions the system understands.
// It is upserted into the permissions table on startup (idempotent), so this
// Go slice remains the single source of truth for what permissions exist.
var Catalog = []CatalogEntry{
	{PermProductsView, "products", "View products"},
	{PermProductsCreate, "products", "Create products"},
	{PermProductsEdit, "products", "Edit products"},
	{PermProductsDelete, "products", "Delete products"},
	{PermProductsRestore, "products", "Restore deleted products"},

	{PermCategoriesView, "categories", "View categories"},
	{PermCategoriesCreate, "categories", "Create categories"},
	{PermCategoriesEdit, "categories", "Edit categories"},
	{PermCategoriesDelete, "categories", "Delete categories"},

	{PermOrdersView, "orders", "View orders"},
	{PermOrdersCreate, "orders", "Create orders"},
	{PermOrdersEdit, "orders", "Edit orders"},
	{PermOrdersCancel, "orders", "Cancel orders"},
	{PermOrdersRefund, "orders", "Refund orders"},
	{PermOrdersPay, "orders", "Take payment for orders"},
	{PermOrdersPrint, "orders", "Print order invoices"},

	{PermCustomersView, "customers", "View customers"},
	{PermCustomersCreate, "customers", "Create customers"},
	{PermCustomersEdit, "customers", "Edit customers"},
	{PermCustomersDelete, "customers", "Delete customers"},

	{PermPromotionsView, "promotions", "View promotions"},
	{PermPromotionsCreate, "promotions", "Create promotions"},
	{PermPromotionsEdit, "promotions", "Edit promotions"},
	{PermPromotionsDelete, "promotions", "Delete promotions"},

	{PermReportsView, "reports", "View reports and dashboards"},

	{PermShiftsManage, "shifts", "Open, close and manage shifts"},

	{PermUsersView, "users", "View staff accounts"},
	{PermUsersCreate, "users", "Create staff accounts"},
	{PermUsersEdit, "users", "Edit staff accounts"},
	{PermUsersDelete, "users", "Delete staff accounts"},

	{PermRolesView, "roles", "View roles"},
	{PermRolesManage, "roles", "Create and edit roles and their permissions"},

	{PermShopsView, "shops", "View shops"},
	{PermShopsManage, "shops", "Create and manage shops"},

	{PermSettingsView, "settings", "View settings"},
	{PermSettingsEdit, "settings", "Edit settings"},

	{PermActivityLogsView, "activity_logs", "View activity logs"},
}

// AllPermissionNames returns every permission name in the catalog.
func AllPermissionNames() []string {
	names := make([]string, 0, len(Catalog))
	for _, e := range Catalog {
		names = append(names, e.Name)
	}
	return names
}
