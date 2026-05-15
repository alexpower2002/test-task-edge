package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type ContractsConfig struct {
	AavePool         types.Address `json:"aave_pool"`
	AaveDataProvider types.Address `json:"aave_data_provider"`
	AaveOracle       types.Address `json:"aave_oracle"`
	MorphoAddress    types.Address `json:"morpho_address"`
	MorphoDeployBlock uint64       `json:"morpho_deploy_block"`
}

type JobConfig struct {
	Network      string          `json:"network"`
	RPCURL       string          `json:"rpc_url"`
	Wallets      []types.Address `json:"wallets"`
	PollInterval time.Duration   `json:"poll_interval"`
	Contracts    ContractsConfig `json:"contracts"`
}

type jobsFile struct {
	Jobs []jobEntry `json:"jobs"`
}

type jobEntry struct {
	Network      string              `json:"network"`
	RPCURL       string              `json:"rpc_url"`
	Wallets      []string            `json:"wallets"`
	PollInterval string              `json:"poll_interval"`
	Contracts    contractsEntry      `json:"contracts"`
}

type contractsEntry struct {
	AavePool         string  `json:"aave_pool"`
	AaveDataProvider string  `json:"aave_data_provider"`
	AaveOracle       string  `json:"aave_oracle"`
	MorphoAddress    string  `json:"morpho_address"`
	MorphoDeployBlock *uint64 `json:"morpho_deploy_block"`
}

type Config struct {
	Jobs   []JobConfig
	PGDSN  string
}

func Load() (Config, error) {
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

	return Config{Jobs: jobs, PGDSN: pgDSN}, nil
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

	wallets, err := utils.ParseAddressList(entry.Wallets)
	if err != nil {
		return JobConfig{}, fmt.Errorf("wallets: %w", err)
	}
	if len(wallets) == 0 {
		return JobConfig{}, fmt.Errorf("at least one wallet required")
	}

	contracts, err := parseContracts(entry.Contracts)
	if err != nil {
		return JobConfig{}, fmt.Errorf("contracts: %w", err)
	}

	return JobConfig{
		Network:      entry.Network,
		RPCURL:       entry.RPCURL,
		Wallets:      wallets,
		PollInterval: pollInterval,
		Contracts:    contracts,
	}, nil
}

func parseContracts(entry contractsEntry) (ContractsConfig, error) {
	if entry.AavePool == "" {
		return ContractsConfig{}, fmt.Errorf("aave_pool is required")
	}
	if entry.AaveDataProvider == "" {
		return ContractsConfig{}, fmt.Errorf("aave_data_provider is required")
	}
	if entry.AaveOracle == "" {
		return ContractsConfig{}, fmt.Errorf("aave_oracle is required")
	}
	if entry.MorphoAddress == "" {
		return ContractsConfig{}, fmt.Errorf("morpho_address is required")
	}
	if entry.MorphoDeployBlock == nil {
		return ContractsConfig{}, fmt.Errorf("morpho_deploy_block is required")
	}

	pool, err := utils.ParseAddress(entry.AavePool)
	if err != nil {
		return ContractsConfig{}, fmt.Errorf("aave_pool: %w", err)
	}
	dataProvider, err := utils.ParseAddress(entry.AaveDataProvider)
	if err != nil {
		return ContractsConfig{}, fmt.Errorf("aave_data_provider: %w", err)
	}
	oracle, err := utils.ParseAddress(entry.AaveOracle)
	if err != nil {
		return ContractsConfig{}, fmt.Errorf("aave_oracle: %w", err)
	}
	morphoAddr, err := utils.ParseAddress(entry.MorphoAddress)
	if err != nil {
		return ContractsConfig{}, fmt.Errorf("morpho_address: %w", err)
	}

	return ContractsConfig{
		AavePool:          pool,
		AaveDataProvider:  dataProvider,
		AaveOracle:        oracle,
		MorphoAddress:     morphoAddr,
		MorphoDeployBlock: *entry.MorphoDeployBlock,
	}, nil
}

func getEnv(key string) (string, error) {
	s := os.Getenv(key)
	if strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return strings.TrimSpace(s), nil
}
