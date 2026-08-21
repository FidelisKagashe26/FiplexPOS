// promote-superadmin repairs the initial platform account without running the
// demo-data seeder. Run from backend/: go run ./cmd/promote-superadmin
package main

import (
	"POS-fiplex/config"
	user_repo "POS-fiplex/internal/user/repository"
	"POS-fiplex/pkg/database"
	"POS-fiplex/pkg/logger"
	"POS-fiplex/sqlc/migrations"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to load .env: %v", err)
	}
	cfg := config.Load()
	logr := logger.New(cfg)
	db, err := database.NewDatabase(cfg, logr, migrations.FS)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := user_repo.New(db.GetPool())
	user, err := repo.GetUserByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Fatalf("Could not find admin@example.com: %v", err)
	}
	name := "Super Admin"
	role := user_repo.UserRoleSuperadmin
	if _, err := repo.UpdateUser(ctx, user_repo.UpdateUserParams{
		ID:       user.ID,
		Username: &name,
		Role:     &role,
	}); err != nil {
		log.Fatalf("Could not promote account: %v", err)
	}
	log.Println("admin@example.com is now Super Admin (superadmin).")
}
