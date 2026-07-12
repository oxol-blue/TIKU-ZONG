package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// This command creates or promotes the configured bootstrap administrator.
// Credentials are read from environment variables and are never printed.
func main() {
	_ = godotenv.Load()
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD are required")
	}
	cfg := config.Load()
	db, err := database.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("MYSQL_DSN is not configured")
	}
	defer db.Close()
	store := auth.NewStore(db)
	ctx := context.Background()
	user, _, err := store.GetUserByEmail(ctx, email)
	if errors.Is(err, auth.ErrNotFound) {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Fatal(hashErr)
		}
		user, err = store.CreateUser(ctx, email, string(hash))
	}
	if err != nil {
		log.Fatal(err)
	}
	if err := store.SetRole(ctx, user.ID, auth.RoleAdmin); err != nil {
		log.Fatal(err)
	}
	log.Printf("administrator ready for user id %d", user.ID)
}
