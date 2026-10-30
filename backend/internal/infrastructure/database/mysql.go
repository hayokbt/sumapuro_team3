package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var schema embed.FS

const setupTimeout = 10 * time.Second

func Open(ctx context.Context) (*sql.DB, error) {
	port, err := strconv.Atoi(envOrDefault("DB_PORT", "3305"))
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("DB_PORT must be a valid TCP port")
	}

	config := mysql.NewConfig()
	config.User = envOrDefault("DB_USER", "gomoku_user")
	config.Passwd = envOrDefault("DB_PASSWORD", "gomoku_password")
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(envOrDefault("DB_HOST", "localhost"), strconv.Itoa(port))
	config.DBName = envOrDefault("DB_NAME", "gomoku_db")
	config.ParseTime = true

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("configure MySQL connection: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)

	setupCtx, cancel := context.WithTimeout(ctx, setupTimeout)
	defer cancel()

	if err := db.PingContext(setupCtx); err != nil {
		return nil, errors.Join(fmt.Errorf("connect to MySQL: %w", err), db.Close())
	}
	if err := initializeSchema(setupCtx, db); err != nil {
		return nil, errors.Join(fmt.Errorf("initialize MySQL schema: %w", err), db.Close())
	}

	return db, nil
}

func initializeSchema(ctx context.Context, db *sql.DB) error {
	statements, err := schema.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	for _, statement := range strings.Split(string(statements), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("execute schema statement: %w", err)
		}
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'auth_version'").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.ExecContext(ctx, "ALTER TABLE users ADD COLUMN auth_version BIGINT NOT NULL DEFAULT 1"); err != nil {
			return err
		}
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
