package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
)

// OpenMySQL creates a MySQL 5.7-compatible connection pool.
func OpenMySQL(cfg config.Config) (*sql.DB, error) {
	if cfg.MySQLDSN == "" {
		return nil, nil
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MySQLMaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQLMaxIdleConns)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
