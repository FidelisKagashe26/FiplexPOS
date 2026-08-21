package seeder

import (
	user_repo "POS-fiplex/internal/user/repository"
	"POS-fiplex/pkg/logger"
	"POS-fiplex/pkg/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SeedUsers creates the initial platform account plus demonstration shop users.
// The display name is Super Admin; email remains the login identity.
func SeedUsers(ctx context.Context, q user_repo.Querier, log logger.ILogger) error {
	userData := []struct {
		Username string
		Email    string
		Role     user_repo.UserRole
	}{
		{"Super Admin", "admin@example.com", user_repo.UserRoleSuperadmin},
		{"manager", "manager@example.com", user_repo.UserRoleManager},
		{"cashier", "cashier@example.com", user_repo.UserRoleCashier},
	}

	hashPassword, err := utils.HashPassword("passwordrahasia")
	if err != nil {
		return err
	}

	for _, data := range userData {
		existing, lookupErr := q.GetUserByEmail(ctx, data.Email)
		if lookupErr == nil {
			// Repair older development seeds, including the display name and role.
			if _, err := q.UpdateUser(ctx, user_repo.UpdateUserParams{
				ID: existing.ID, Username: &data.Username, Role: &data.Role,
			}); err != nil {
				return err
			}
			continue
		}
		if !errors.Is(lookupErr, pgx.ErrNoRows) {
			return lookupErr
		}

		userUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		_, err = q.CreateUser(ctx, user_repo.CreateUserParams{
			ID: userUUID, Username: data.Username, Email: data.Email,
			PasswordHash: hashPassword, IsActive: true, Role: data.Role,
		})
		if err != nil {
			return err
		}
		log.Infof("Seeder User | created %s as %s", data.Email, data.Role)
	}
	return nil
}
