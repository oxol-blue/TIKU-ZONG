package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ai"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/announcement"
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
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/settings"
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
	var announcementService *announcement.Service
	var settingsService *settings.Service
	captchaStore := captcha.NewStore()
	if db != nil {
		authService = auth.NewService(auth.NewStore(db), cfg.JWTSecret, captchaStore)
		questionService = questions.NewService(questions.NewStore(db))
		billingService = billing.NewService(billing.NewStore(db))
		callLogger = calls.NewStore(db)
		aiService = ai.NewServiceWithQueue(ai.NewStore(db, cfg.EncryptionSecret), cfg.AIQueueSize, cfg.AIQueueWorkers)
		ocsStore = ocs.NewStore(db)
		ocsService = ocs.NewService(ocsStore, cfg.AnswerMergeRule)
		paymentService = payment.NewService(payment.NewStore(db, cfg.EncryptionSecret), cfg.PublicBaseURL)
		feedbackService = feedback.NewService(feedback.NewStore(db))
		announcementService = announcement.NewService(announcement.NewStore(db))
		settingsService = settings.NewService(settings.NewStore(db))
	}
	router := httpapi.NewRouter(cfg, authService, questionService, billingService, callLogger, aiService, ocsStore, db, paymentService, feedbackService, ocsService, announcementService, settingsService)
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
	reconciliationTicker := time.NewTicker(5 * time.Minute)
	defer reconciliationTicker.Stop()
	for range ticker.C {
		closed, err := service.Store().CloseExpired(context.Background(), time.Now().UTC())
		if err != nil {
			log.Printf("payment maintenance failed: %v", err)
			continue
		}
		if closed > 0 {
			log.Printf("payment maintenance closed %d expired order(s)", closed)
		}
		select {
		case <-reconciliationTicker.C:
			issues, reconcileErr := service.Store().Reconcile(context.Background(), time.Now().UTC())
			if reconcileErr != nil {
				log.Printf("payment reconciliation failed: %v", reconcileErr)
			} else if len(issues) > 0 {
				log.Printf("payment reconciliation found %d issue(s)", len(issues))
			}
			repaired, repairErr := service.Store().RepairMissingPackageInstances(context.Background())
			if repairErr != nil {
				log.Printf("payment package-instance repair failed: %v", repairErr)
			} else if repaired > 0 {
				log.Printf("payment package-instance repair restored %d order(s)", repaired)
			}
		default:
		}
	}
}
