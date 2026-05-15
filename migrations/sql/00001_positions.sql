-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    protocol VARCHAR(64) NOT NULL,
    network VARCHAR(32) NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    market_id VARCHAR(128) NOT NULL DEFAULT '',
    collateral_token VARCHAR(64) NOT NULL DEFAULT '',
    debt_token VARCHAR(64) NOT NULL DEFAULT '',
    position_size NUMERIC NOT NULL DEFAULT 0,
    token_price NUMERIC NOT NULL DEFAULT 0,
    health_factor NUMERIC NOT NULL DEFAULT 0,
    block_number BIGINT NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS positions;
-- +goose StatementEnd
