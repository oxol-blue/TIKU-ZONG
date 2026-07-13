package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/billing"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/calls"
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
	var billingService *billing.Service
	var callLogger *calls.Store
	if db != nil {
		authService = auth.NewService(auth.NewStore(db), cfg.JWTSecret)
		questionService = questions.NewService(questions.NewStore(db))
		billingService = billing.NewService(billing.NewStore(db))
		callLogger = calls.NewStore(db)
	}
	router := httpapi.NewRouter(cfg, authService, questionService, billingService, callLogger)

	log.Printf("%s starting on %s", cfg.AppName, cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
