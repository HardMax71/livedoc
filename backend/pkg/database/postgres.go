package database

import (
	"context"
	"fmt"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/config"
	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	logger := utils.Logger()

	connString := cfg.GetConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	poolConfig.MaxConns = 50
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	logger.Info("Successfully connected to database")

	return &Database{pool: pool}, nil
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

// Transaction executes the given function within a transaction
func (db *Database) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
