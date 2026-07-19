package main

import (
	"POS-fiplex/config"
	cancellation_reasons_repo "POS-fiplex/internal/cancellation_reasons/repository"
	categories_repo "POS-fiplex/internal/categories/repository"
	payment_methods_repo "POS-fiplex/internal/payment_methods/repository"
	user_repo "POS-fiplex/internal/user/repository"
	cloudflarer2 "POS-fiplex/pkg/cloudflare-r2"
	"POS-fiplex/pkg/database"
	"POS-fiplex/pkg/database/seeder"
	"POS-fiplex/pkg/logger"
	"POS-fiplex/sqlc/migrations"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	envFile := ".env"
	if len(os.Args) > 1 {
		envFile = os.Args[1]
	}

	err := godotenv.Load(envFile)
	if err != nil {
		// .env lives at the repository root when running from backend/
		if err = godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: Error loading %s file: %v", envFile, err)
		}
	}

	cfg := config.Load()

	ctx := context.Background()

	logr := logger.New(cfg)

	db, err := database.NewDatabase(cfg, logr, migrations.FS)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	pool := db.GetPool()

	r2Client, err := cloudflarer2.NewCloudflareR2(cfg, logr)
	if err != nil {
		log.Printf("Failed to initialize R2 client (images will not be uploaded): %v", err)
	}

	userRepo := user_repo.New(pool)
	catRepo := categories_repo.New(pool)
	paymentMethodRepo := payment_methods_repo.New(pool)
	cancelRepo := cancellation_reasons_repo.New(pool)

	if err := seeder.RunSeeders(ctx, pool, userRepo, catRepo, paymentMethodRepo, cancelRepo, r2Client, logr); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully.")
}
