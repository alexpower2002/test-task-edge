package protocol

type Position struct {
	Protocol        string `json:"protocol"`
	Network         string `json:"network"`
	WalletAddress   string `json:"wallet_address"`
	MarketID        string `json:"market_id"`
	CollateralToken string `json:"collateral_token"`
	DebtToken       string `json:"debt_token"`
	PositionSize    string `json:"position_size"`
	TokenPrice      string `json:"token_price"`
	HealthFactor    string `json:"health_factor"`
	BlockNumber     uint64 `json:"block_number"`
	Timestamp       uint64 `json:"timestamp"`
}
