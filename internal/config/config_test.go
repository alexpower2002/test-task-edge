package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullContracts = `"contracts": {
	"aave_pool": "0x1111111111111111111111111111111111111111",
	"aave_data_provider": "0x2222222222222222222222222222222222222222",
	"aave_oracle": "0x3333333333333333333333333333333333333333",
	"morpho_address": "0x4444444444444444444444444444444444444444",
	"morpho_deploy_block": 100
}`

func writeJobsFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "jobs.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		jobsJSON     string
		env          map[string]string
		noFile       bool
		wantErr      bool
		wantJobs     int
		wantNetwork  string
		wantRPC      string
		wantWallets  int
		wantPool     string
		wantInterval time.Duration
	}{
		{
			name: "ok_single_job",
			jobsJSON: `{
				"jobs": [{
					"network": "ethereum",
					"rpc_url": "http://localhost:8545",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					"poll_interval": "2s",
					` + fullContracts + `
				}]
			}`,
			env:         map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantJobs:    1,
			wantNetwork: "ethereum",
			wantRPC:     "http://localhost:8545",
			wantWallets: 1,
			wantPool:    "0x1111111111111111111111111111111111111111",
			wantInterval: 2 * time.Second,
		},
		{
			name: "ok_multiple_jobs",
			jobsJSON: `{
				"jobs": [
					{
						"network": "ethereum",
						"rpc_url": "http://eth-rpc",
						"wallets": ["0x0000000000000000000000000000000000000001"],
						"poll_interval": "1s",
						` + fullContracts + `
					},
					{
						"network": "arbitrum",
						"rpc_url": "http://arb-rpc",
						"wallets": ["0x0000000000000000000000000000000000000002"],
						"poll_interval": "1s",
						` + fullContracts + `
					}
				]
			}`,
			env:         map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantJobs:    2,
			wantNetwork: "ethereum",
			wantRPC:     "http://eth-rpc",
			wantWallets: 1,
			wantPool:    "0x1111111111111111111111111111111111111111",
		},
		{
			name: "err_missing_pg_dsn",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					"poll_interval": "1s",
					` + fullContracts + `
				}]
			}`,
			wantErr: true,
		},
		{
			name:    "err_missing_jobs_file",
			env:     map[string]string{"PG_DSN": "postgres://localhost/test"},
			noFile:  true,
			wantErr: true,
		},
		{
			name: "err_missing_poll_interval",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					` + fullContracts + `
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_contracts",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					"poll_interval": "1s"
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_wallets",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"wallets": [],
					"poll_interval": "1s",
					` + fullContracts + `
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_no_jobs",
			jobsJSON: `{
				"jobs": []
			}`,
			wantErr: true,
		},
		{
			name: "err_bad_poll_interval",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					"poll_interval": "bad",
					` + fullContracts + `
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_network",
			jobsJSON: `{
				"jobs": [{
					"rpc_url": "http://rpc",
					"wallets": ["0x0000000000000000000000000000000000000001"],
					"poll_interval": "1s",
					` + fullContracts + `
				}]
			}`,
			wantErr: true,
		},
		{
			name:    "err_invalid_json",
			jobsJSON: `{bad`,
			env:     map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.noFile {
				path := writeJobsFile(t, tt.jobsJSON)
				t.Setenv("JOBS_FILE", path)
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, cfg.Jobs, tt.wantJobs)
			job := cfg.Jobs[0]
			assert.Equal(t, tt.wantNetwork, job.Network)
			assert.Equal(t, tt.wantRPC, job.RPCURL)
			assert.Equal(t, tt.wantWallets, len(job.Wallets))
			assert.Equal(t, tt.wantPool, job.Contracts.AavePool.String())
			if tt.wantInterval > 0 {
				assert.Equal(t, tt.wantInterval, job.PollInterval)
			}
		})
	}
}
