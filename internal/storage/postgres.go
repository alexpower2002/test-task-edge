package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"test-task-edge/internal/protocol"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Connect: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("goose.SetDialect: %w", err)
	}
	migrationsDir := "migrations/sql"
	if err := goose.Up(db.DB, migrationsDir); err != nil {
		return nil, fmt.Errorf("goose.Up: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (s *Postgres) Close() error {
	return s.db.Close()
}

func (s *Postgres) SavePositions(ctx context.Context, positions []protocol.Position) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO positions
			(protocol, network, wallet_address, market_id, collateral_token, debt_token,
			 position_size, token_price, health_factor, block_number, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, p := range positions {
		if _, err := stmt.ExecContext(ctx,
			p.Protocol, p.Network, p.WalletAddress, p.MarketID, p.CollateralToken, p.DebtToken,
			p.PositionSize, p.TokenPrice, p.HealthFactor, p.BlockNumber, p.Timestamp,
		); err != nil {
			return fmt.Errorf("insert position: %w", err)
		}
	}

	return tx.Commit()
}
