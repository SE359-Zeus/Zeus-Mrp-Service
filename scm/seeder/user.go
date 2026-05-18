package seeder

import (
	"time"
	"github.com/google/uuid"
	"github.com/brianvoe/gofakeit/v6"
	"gorm.io/gorm"
	"zeus-scm-service/internal/models"
)

func seedUsers(db *gorm.DB, count int) []models.User {
	var users []models.User
	for i := 0; i < count; i++ {
		u := models.User{
			ID:            uuid.New(),
			AccountStatus: 1,
			RoleID:        1,
			Email:         gofakeit.Email(),
			FullName:      gofakeit.Name(),
			PasswordHash:  "fakehash",
			PhoneNumber:   gofakeit.Phone(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		db.Create(&u)
		users = append(users, u)
	}
	return users
}
