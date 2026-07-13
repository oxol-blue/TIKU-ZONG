package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ai"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/billing"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/calls"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/captcha"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/database"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/feedback"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/httpapi"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ocs"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/payment"
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
	var aiService *ai.Service
	var ocsStore *ocs.Store
	var ocsService *ocs.Service
	var paymentService *payment.Service
	var feedbackService *feedback.Service
	captchaStore := captcha.NewStore()
	if db != nil {
		authService = auth.NewService(auth.NewStore(db), cfg.JWTSecret, captchaStore)
		questionService = questions.NewService(questions.NewStore(db))
		billingService = billing.NewService(billing.NewStore(db))
		callLogger = calls.NewStore(db)
		aiService = ai.NewService(ai.NewStore(db, cfg.EncryptionSecret))
		ocsStore = ocs.NewStore(db)
		ocsService = ocs.NewService(ocsStore, cfg.AnswerMergeRule)
		paymentService = payment.NewService(payment.NewStore(db, cfg.EncryptionSecret), cfg.PublicBaseURL)
		feedbackService = feedback.NewService(feedback.NewStore(db))
	}
	router := httpapi.NewRouter(cfg, authService, questionService, billingService, callLogger, aiService, ocsStore, paymentService, feedbackService, ocsService)
	if paymentService != nil {
		go runPaymentMaintenance(paymentService)
	}

	log.Printf("%s starting on %s", cfg.AppName, cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}

func runPaymentMaintenance(service *payment.Service) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		closed, err := service.Store().CloseExpired(context.Background(), time.Now().UTC())
		if err != nil {
			log.Printf("payment maintenance failed: %v", err)
			continue
		}
		if closed > 0 {
			log.Printf("payment maintenance closed %d expired order(s)", closed)
		}
	}
}
