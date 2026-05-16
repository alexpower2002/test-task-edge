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

	aaveReserveDiscoverer := aavev3.NewReserveDiscoverer(client, job.Aave.DataProvider)
	aaveUserDataReader := aavev3.NewUserDataReader(client, job.Aave.DataProvider)
	aaveAssetPricer := aavev3.NewAssetPricer(client, job.Aave.Oracle)
	aaveHealthReader := aavev3.NewHealthFactorReader(client, job.Aave.Pool)

	morphoPositionReader := morpho.NewPositionReader(client, job.Morpho.Address)
	morphoMarketParamsReader := morpho.NewMarketParamsReader(client, job.Morpho.Address)
	morphoMarketDataReader := morpho.NewMarketDataReader(client, job.Morpho.Address)
	morphoOraclePricer := morpho.NewOraclePricer(client)
	morphoHealthComputer := morpho.NewHealthFactorReader()

	disc := morpho.NewMarketDiscoverer(client, job.Morpho.Address, job.Morpho.DeployBlock)
	provider := morpho.NewMarketProvider(disc)
	parser := protocol.NewComposite(
		protocol.SubParser{Parser: aavev3.NewParser(job.Aave.Pool, tokenProvider, aaveReserveDiscoverer, aaveUserDataReader, aaveAssetPricer, aaveHealthReader), Wallets: job.AaveWallets},
		protocol.SubParser{Parser: morpho.NewParser(tokenProvider, provider, morphoPositionReader, morphoMarketParamsReader, morphoMarketDataReader, morphoOraclePricer, morphoHealthComputer, job.Morpho.Parallelism), Wallets: job.MorphoWallets},
	)
	a.workers = append(a.workers, worker.New(job.Network, client, parser, nil, a.storage))

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
