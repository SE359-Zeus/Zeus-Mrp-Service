package seeder

import (
	"log"
	"time"

	"zeus-system-service/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAll(db *gorm.DB) error {
	log.Println("Seeding system service data...")

	admin := seedUser(db, "admin@zeus.com", "admin123", "System Administrator", models.UserRoleAdmin)
	editor := seedUser(db, "editor@zeus.com", "editor123", "Production Editor", models.UserRoleEditor)
	viewer := seedUser(db, "viewer@zeus.com", "viewer123", "Read-Only Viewer", models.UserRoleViewer)

	seedAuditLogs(db, admin, editor, viewer)

	log.Println("Seeding complete.")
	return nil
}

func seedUser(db *gorm.DB, email, password, fullName string, role models.UserRole) *models.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password for %s: %v", email, err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         role,
		Status:       models.AccountStatusActive,
	}

	result := db.Where("email = ?", email).FirstOrCreate(user)
	if result.Error != nil {
		log.Fatalf("failed to seed user %s: %v", email, result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Created user: %s (%s) — %s", email, string(role), fullName)
	} else {
		log.Printf("User already exists: %s", email)
	}

	return user
}

func seedAuditLogs(db *gorm.DB, users ...*models.User) {
	now := time.Now()

	type seedEvent struct {
		user            *models.User
		action          models.ActionType
		target          string
		details         string
		ip              string
		isSecurity      bool
		hoursAgo        int
	}

	events := []seedEvent{
		{users[0], models.ActionLogin, "auth/login", "Successful login", "10.0.0.1", false, 1},
		{users[0], models.ActionLogin, "auth/login", "Successful login", "10.0.0.1", false, 2},
		{users[1], models.ActionLogin, "auth/login", "Successful login", "10.0.0.2", false, 3},
		{users[0], models.ActionCreate, "users/" + users[1].ID.String(), "Created editor account", "10.0.0.1", false, 4},
		{users[0], models.ActionCreate, "users/" + users[2].ID.String(), "Created viewer account", "10.0.0.1", false, 5},
		{users[1], models.ActionUpdate, "users/" + users[2].ID.String(), "Updated user role", "10.0.0.2", false, 6},
		{users[0], models.ActionUpdate, "users/" + users[1].ID.String(), "Updated user profile", "10.0.0.1", false, 7},
		{users[0], models.ActionDelete, "users/old-user-id", "Removed inactive account", "10.0.0.1", false, 8},
		{users[0], models.ActionSecurity, "auth/login", "Failed login attempt from unknown IP", "203.0.113.1", true, 9},
		{users[0], models.ActionSecurity, "auth/login", "Brute force attempt detected", "198.51.100.1", true, 10},
		{users[1], models.ActionCreate, "reports/Q1-2026", "Created quarterly report", "10.0.0.2", false, 11},
		{users[0], models.ActionUpdate, "config/security", "Updated password policy", "10.0.0.1", false, 12},
		{users[0], models.ActionLogin, "auth/login", "Successful login", "10.0.0.1", false, 13},
		{users[0], models.ActionCreate, "roles/custom-role", "Created custom audit role", "10.0.0.1", false, 14},
		{users[1], models.ActionUpdate, "reports/Q1-2026", "Updated report filters", "10.0.0.2", false, 15},
		{users[0], models.ActionSecurity, "resources/confidential", "Unauthorized access attempt", "192.0.2.1", true, 16},
	}

	for _, e := range events {
		ts := now.Add(-time.Duration(e.hoursAgo) * time.Hour)
		logEntry := &models.AuditLog{
			UserID:          e.user.ID,
			UserEmail:       e.user.Email,
			ActionType:      e.action,
			TargetResource:  e.target,
			Details:         e.details,
			IPAddress:       e.ip,
			IsSecurityEvent: e.isSecurity,
			Timestamp:       ts,
		}

		result := db.Create(logEntry)
		if result.Error != nil {
			log.Printf("Warning: failed to create audit log entry: %v", result.Error)
		}
	}

	log.Printf("Seeded %d audit log entries", len(events))
}

func init() {
	uuid.New() // ensure uuid package is linked
}
