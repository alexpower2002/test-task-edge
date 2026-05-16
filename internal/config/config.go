package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type AaveConfig struct {
	Pool         types.Address `json:"pool"`
	DataProvider types.Address `json:"data_provider"`
	Oracle       types.Address `json:"oracle"`
}

type MorphoConfig struct {
	Address     types.Address `json:"address"`
	DeployBlock uint64        `json:"deploy_block"`
	Parallelism int           `json:"parallelism"`
}

type JobConfig struct {
	Network      string          `json:"network"`
	RPCURL       string          `json:"rpc_url"`
	PollInterval time.Duration   `json:"poll_interval"`
	Aave         AaveConfig      `json:"aave"`
	Morpho       MorphoConfig    `json:"morpho"`
	AaveWallets  []types.Address `json:"-"`
	MorphoWallets []types.Address `json:"-"`
}

type jobsFile struct {
	Jobs []jobEntry `json:"jobs"`
}

type aaveEntry struct {
	Pool         string   `json:"pool"`
	DataProvider string   `json:"data_provider"`
	Oracle       string   `json:"oracle"`
	Wallets      []string `json:"wallets"`
}

type morphoEntry struct {
	Address     string   `json:"address"`
	DeployBlock *uint64  `json:"deploy_block"`
	Wallets     []string `json:"wallets"`
	Parallelism *int     `json:"parallelism,omitempty"`
}

type jobEntry struct {
	Network      string        `json:"network"`
	RPCURL       string        `json:"rpc_url"`
	PollInterval string        `json:"poll_interval"`
	Aave         *aaveEntry    `json:"aave"`
	Morpho       *morphoEntry  `json:"morpho"`
}

type Config struct {
	Jobs          []JobConfig
	PGDSN         string
	RPCMaxRetries int
	RPCRetryWait  time.Duration
	RPCTimeout    time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	pgDSN, err := getEnv("PG_DSN")
	if err != nil {
		return Config{}, err
	}

	path, err := getEnv("JOBS_FILE")
	if err != nil {
		return Config{}, err
	}
	jobs, err := loadJobs(path)
	if err != nil {
		return Config{}, fmt.Errorf("load jobs: %w", err)
	}
	if len(jobs) == 0 {
		return Config{}, fmt.Errorf("at least one job is required")
	}

	rpcMaxRetries := 3
	rpcRetryWait := time.Second
	rpcTimeout := 15 * time.Second

	if v := os.Getenv("RPC_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err != nil {
			return Config{}, fmt.Errorf("invalid RPC_MAX_RETRIES: %q", v)
		} else {
			rpcMaxRetries = n
		}
	}
	if v := os.Getenv("RPC_RETRY_WAIT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid RPC_RETRY_WAIT: %w", err)
		}
		rpcRetryWait = d
	}
	if v := os.Getenv("RPC_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid RPC_TIMEOUT: %w", err)
		}
		rpcTimeout = d
	}

	return Config{
		Jobs:          jobs,
		PGDSN:         pgDSN,
		RPCMaxRetries: rpcMaxRetries,
		RPCRetryWait:  rpcRetryWait,
		RPCTimeout:    rpcTimeout,
	}, nil
}

func loadJobs(path string) ([]JobConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file jobsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	jobs := make([]JobConfig, 0, len(file.Jobs))
	for _, entry := range file.Jobs {
		job, err := parseJob(entry)
		if err != nil {
			return nil, fmt.Errorf("job %q: %w", entry.Network, err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func parseJob(entry jobEntry) (JobConfig, error) {
	if entry.Network == "" {
		return JobConfig{}, fmt.Errorf("network is required")
	}
	if entry.RPCURL == "" {
		return JobConfig{}, fmt.Errorf("rpc_url is required")
	}
	if entry.PollInterval == "" {
		return JobConfig{}, fmt.Errorf("poll_interval is required")
	}

	pollInterval, err := time.ParseDuration(strings.TrimSpace(entry.PollInterval))
	if err != nil {
		return JobConfig{}, fmt.Errorf("invalid poll_interval: %w", err)
	}

	if entry.Aave == nil {
		return JobConfig{}, fmt.Errorf("aave config is required")
	}
	if entry.Morpho == nil {
		return JobConfig{}, fmt.Errorf("morpho config is required")
	}

	aaveCfg, err := parseAave(*entry.Aave)
	if err != nil {
		return JobConfig{}, fmt.Errorf("aave: %w", err)
	}

	morphoCfg, err := parseMorpho(*entry.Morpho)
	if err != nil {
		return JobConfig{}, fmt.Errorf("morpho: %w", err)
	}

	aaveWallets, err := utils.ParseAddressList(entry.Aave.Wallets)
	if err != nil {
		return JobConfig{}, fmt.Errorf("aave wallets: %w", err)
	}
	if len(aaveWallets) == 0 {
		return JobConfig{}, fmt.Errorf("at least one aave wallet required")
	}

	morphoWallets, err := utils.ParseAddressList(entry.Morpho.Wallets)
	if err != nil {
		return JobConfig{}, fmt.Errorf("morpho wallets: %w", err)
	}
	if len(morphoWallets) == 0 {
		return JobConfig{}, fmt.Errorf("at least one morpho wallet required")
	}

	return JobConfig{
		Network:       entry.Network,
		RPCURL:        entry.RPCURL,
		PollInterval:  pollInterval,
		Aave:          aaveCfg,
		Morpho:        morphoCfg,
		AaveWallets:   aaveWallets,
		MorphoWallets: morphoWallets,
	}, nil
}

func parseAave(entry aaveEntry) (AaveConfig, error) {
	if entry.Pool == "" {
		return AaveConfig{}, fmt.Errorf("pool is required")
	}
	if entry.DataProvider == "" {
		return AaveConfig{}, fmt.Errorf("data_provider is required")
	}
	if entry.Oracle == "" {
		return AaveConfig{}, fmt.Errorf("oracle is required")
	}

	pool, err := utils.ParseAddress(entry.Pool)
	if err != nil {
		return AaveConfig{}, fmt.Errorf("pool: %w", err)
	}
	dataProvider, err := utils.ParseAddress(entry.DataProvider)
	if err != nil {
		return AaveConfig{}, fmt.Errorf("data_provider: %w", err)
	}
	oracle, err := utils.ParseAddress(entry.Oracle)
	if err != nil {
		return AaveConfig{}, fmt.Errorf("oracle: %w", err)
	}

	return AaveConfig{
		Pool:         pool,
		DataProvider: dataProvider,
		Oracle:       oracle,
	}, nil
}

func parseMorpho(entry morphoEntry) (MorphoConfig, error) {
	if entry.Address == "" {
		return MorphoConfig{}, fmt.Errorf("address is required")
	}
	if entry.DeployBlock == nil {
		return MorphoConfig{}, fmt.Errorf("deploy_block is required")
	}

	addr, err := utils.ParseAddress(entry.Address)
	if err != nil {
		return MorphoConfig{}, fmt.Errorf("address: %w", err)
	}

	parallelism := 100
	if entry.Parallelism != nil {
		parallelism = *entry.Parallelism
	}

	return MorphoConfig{
		Address:     addr,
		DeployBlock: *entry.DeployBlock,
		Parallelism: parallelism,
	}, nil
}

func getEnv(key string) (string, error) {
	s := os.Getenv(key)
	if strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return strings.TrimSpace(s), nil
}
