package lendingparser

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/config"
	"test-task-edge/internal/ethereum"
	"test-task-edge/internal/protocol"
	"test-task-edge/internal/protocol/aavev3"
	"test-task-edge/internal/protocol/morpho"
	"test-task-edge/internal/storage"
	"test-task-edge/internal/worker"
)

type app struct {
	clients []*ethereum.RPCClient
	workers []*worker.Worker
	storage *storage.Postgres
}

func NewApp() *app {
	return &app{}
}

func (a *app) Register() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	a.storage, err = storage.NewPostgres(cfg.PGDSN)
	if err != nil {
		return err
	}

	for _, job := range cfg.Jobs {
		if err := a.registerJob(cfg, job); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) registerJob(cfg config.Config, job config.JobConfig) error {
	log.Info().Str("network", job.Network).Msg("registering job")

	client, err := ethereum.NewRPCClient(job.RPCURL, job.PollInterval, cfg.RPCMaxRetries, cfg.RPCRetryWait, cfg.RPCTimeout)
	if err != nil {
		return err
	}
	a.clients = append(a.clients, client)

	tokenReader := ethereum.NewTokenReader(client)
	tokenProvider := ethereum.NewTokenProvider(tokenReader)

	aaveReserveDiscoverer := aavev3.NewReserveDiscoverer(client, job.Contracts.AaveDataProvider)
	aaveUserDataReader := aavev3.NewUserDataReader(client, job.Contracts.AaveDataProvider)
	aaveAssetPricer := aavev3.NewAssetPricer(client, job.Contracts.AaveOracle)
	aaveHealthReader := aavev3.NewHealthFactorReader(client, job.Contracts.AavePool)

	morphoPositionReader := morpho.NewPositionReader(client, job.Contracts.MorphoAddress)
	morphoMarketParamsReader := morpho.NewMarketParamsReader(client, job.Contracts.MorphoAddress)
	morphoMarketDataReader := morpho.NewMarketDataReader(client, job.Contracts.MorphoAddress)
	morphoOraclePricer := morpho.NewOraclePricer(client)
	morphoHealthComputer := morpho.NewHealthFactorReader()

	disc := morpho.NewMarketDiscoverer(client, job.Contracts.MorphoAddress, job.Contracts.MorphoDeployBlock)
	provider := morpho.NewMarketProvider(disc)
	parser := protocol.NewComposite(
		aavev3.NewParser(job.Contracts.AavePool, tokenProvider, aaveReserveDiscoverer, aaveUserDataReader, aaveAssetPricer, aaveHealthReader),
		morpho.NewParser(tokenProvider, provider, morphoPositionReader, morphoMarketParamsReader, morphoMarketDataReader, morphoOraclePricer, morphoHealthComputer),
	)
	a.workers = append(a.workers, worker.New(job.Network, client, parser, job.Wallets, a.storage))

	return nil
}

func (a *app) Resolve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, w := range a.workers {
		wg.Add(1)
		go func(w *worker.Worker) {
			defer wg.Done()
			if err := w.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("worker exited with error")
			}
		}(w)
	}

	wg.Wait()
	return nil
}

func (a *app) Release() error {
	for _, c := range a.clients {
		c.Close()
	}
	if a.storage != nil {
		a.storage.Close()
	}
	return nil
}
