package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db, err := database.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("MYSQL_DSN is not configured")
	}
	defer db.Close()

	if err := ensureMigrationsTable(db); err != nil {
		log.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join("migrations", "*.sql"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)
	for _, file := range files {
		if err := applyMigration(db, file); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("migrations complete: %d file(s)", len(files))
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version BIGINT NOT NULL PRIMARY KEY,
        applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	return err
}

func applyMigration(db *sql.DB, file string) error {
	version, err := migrationVersion(file)
	if err != nil {
		return err
	}
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range strings.Split(string(content), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("applied migration %s", filepath.Base(file))
	return nil
}

func migrationVersion(file string) (int64, error) {
	base := filepath.Base(file)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", base)
	}
	return strconv.ParseInt(parts[0], 10, 64)
}
