package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/database"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/httpapi"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/questions"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("local .env not loaded: %v", err)
	}
	cfg := config.Load()
	db, err := database.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}
	var authService *auth.Service
	var questionService *questions.Service
	if db != nil {
		authService = auth.NewService(auth.NewStore(db), cfg.JWTSecret)
		questionService = questions.NewService(questions.NewStore(db))
	}
	router := httpapi.NewRouter(cfg, authService, questionService)

	log.Printf("%s starting on %s", cfg.AppName, cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
