package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullAave = `"aave": {
	"pool": "0x1111111111111111111111111111111111111111",
	"data_provider": "0x2222222222222222222222222222222222222222",
	"oracle": "0x3333333333333333333333333333333333333333",
	"wallets": ["0x0000000000000000000000000000000000000001"]
}`

const fullMorpho = `"morpho": {
	"address": "0x4444444444444444444444444444444444444444",
	"deploy_block": 100,
	"scan_batch_size": 50000,
	"wallets": ["0x0000000000000000000000000000000000000001"]
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
		name          string
		jobsJSON      string
		env           map[string]string
		noFile        bool
		wantErr       bool
		wantJobs      int
		wantNetwork   string
		wantRPC       string
		wantPool      string
		wantInterval  time.Duration
		wantBatchSize uint64
	}{
		{
			name: "ok_single_job",
			jobsJSON: `{
				"jobs": [{
					"network": "ethereum",
					"rpc_url": "http://localhost:8545",
					"poll_interval": "2s",
					` + fullAave + `,
					` + fullMorpho + `
				}]
			}`,
			env:           map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantJobs:      1,
			wantNetwork:   "ethereum",
			wantRPC:       "http://localhost:8545",
			wantPool:      "0x1111111111111111111111111111111111111111",
			wantInterval:  2 * time.Second,
			wantBatchSize: 50000,
		},
		{
			name: "ok_multiple_jobs",
			jobsJSON: `{
				"jobs": [
					{
						"network": "ethereum",
						"rpc_url": "http://eth-rpc",
						"poll_interval": "1s",
						` + fullAave + `,
						` + fullMorpho + `
					},
					{
						"network": "arbitrum",
						"rpc_url": "http://arb-rpc",
						"poll_interval": "1s",
						"aave": {
							"pool": "0x1111111111111111111111111111111111111111",
							"data_provider": "0x2222222222222222222222222222222222222222",
							"oracle": "0x3333333333333333333333333333333333333333",
							"wallets": ["0x0000000000000000000000000000000000000002"]
						},
						"morpho": {
							"address": "0x4444444444444444444444444444444444444444",
							"deploy_block": 200,
							"wallets": ["0x0000000000000000000000000000000000000002"]
						}
					}
				]
			}`,
			env:         map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantJobs:    2,
			wantNetwork: "ethereum",
			wantRPC:     "http://eth-rpc",
			wantPool:    "0x1111111111111111111111111111111111111111",
		},
		{
			name: "ok_default_scan_batch_size",
			jobsJSON: `{
				"jobs": [{
					"network": "ethereum",
					"rpc_url": "http://localhost:8545",
					"poll_interval": "1s",
					` + fullAave + `,
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": ["0x0000000000000000000000000000000000000001"]
					}
				}]
			}`,
			env:           map[string]string{"PG_DSN": "postgres://localhost/test"},
			wantJobs:      1,
			wantNetwork:   "ethereum",
			wantRPC:       "http://localhost:8545",
			wantPool:      "0x1111111111111111111111111111111111111111",
			wantInterval: time.Second,
			wantBatchSize: 10000,
		},
		{
			name: "err_missing_pg_dsn",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					` + fullAave + `,
					` + fullMorpho + `
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
					"aave": {
						"pool": "0x1111111111111111111111111111111111111111",
						"data_provider": "0x2222222222222222222222222222222222222222",
						"oracle": "0x3333333333333333333333333333333333333333",
						"wallets": ["0x0000000000000000000000000000000000000001"]
					},
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": ["0x0000000000000000000000000000000000000001"]
					}
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_aave",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					` + fullMorpho + `
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_morpho",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					` + fullAave + `
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_aave_wallets",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					"aave": {
						"pool": "0x1111111111111111111111111111111111111111",
						"data_provider": "0x2222222222222222222222222222222222222222",
						"oracle": "0x3333333333333333333333333333333333333333",
						"wallets": []
					},
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": ["0x0000000000000000000000000000000000000001"]
					}
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_morpho_wallets",
			jobsJSON: `{
				"jobs": [{
					"network": "e",
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					"aave": {
						"pool": "0x1111111111111111111111111111111111111111",
						"data_provider": "0x2222222222222222222222222222222222222222",
						"oracle": "0x3333333333333333333333333333333333333333",
						"wallets": ["0x0000000000000000000000000000000000000001"]
					},
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": []
					}
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
					"poll_interval": "bad",
					"aave": {
						"pool": "0x1111111111111111111111111111111111111111",
						"data_provider": "0x2222222222222222222222222222222222222222",
						"oracle": "0x3333333333333333333333333333333333333333",
						"wallets": ["0x0000000000000000000000000000000000000001"]
					},
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": ["0x0000000000000000000000000000000000000001"]
					}
				}]
			}`,
			wantErr: true,
		},
		{
			name: "err_missing_network",
			jobsJSON: `{
				"jobs": [{
					"rpc_url": "http://rpc",
					"poll_interval": "1s",
					"aave": {
						"pool": "0x1111111111111111111111111111111111111111",
						"data_provider": "0x2222222222222222222222222222222222222222",
						"oracle": "0x3333333333333333333333333333333333333333",
						"wallets": ["0x0000000000000000000000000000000000000001"]
					},
					"morpho": {
						"address": "0x4444444444444444444444444444444444444444",
						"deploy_block": 100,
						"wallets": ["0x0000000000000000000000000000000000000001"]
					}
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
			assert.Equal(t, tt.wantPool, job.Aave.Pool.String())
			if tt.wantInterval > 0 {
				assert.Equal(t, tt.wantInterval, job.PollInterval)
			}
			if tt.wantBatchSize > 0 {
				assert.Equal(t, tt.wantBatchSize, job.Morpho.ScanBatchSize)
			}
		})
	}
}
