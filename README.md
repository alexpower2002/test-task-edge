# Парсер lending-позиций

На каждом новом блоке проверяет долговые позиции по списку кошельков.
Протоколы: Aave v3 (обязательный) и Morpho Blue (дополнительный).
Сети: Ethereum и Arbitrum (Arbitrum добавил, чтобы показать, что новая сеть заводится без правки кода).
Цены только из on-chain оракулов.

## Запуск

Берём из примеров `.env` и `jobs.json` и через docker-compose запускаем.

```bash
docker compose up
```

Переменные окружения в `.env`: `JOBS_FILE` (путь к `jobs.json`), `PG_DSN` (строка подключения к Postgres), `RPC_MAX_RETRIES` (сколько раз повторять запрос RPC-эндпоинту, по умолчанию 3), `RPC_RETRY_WAIT` (пауза между повторами, по умолчанию 1с), `RPC_TIMEOUT` (таймаут RPC, по умолчанию 15с).

Конфигурация сетей и протоколов в `jobs.json`:

```json
{
  "jobs": [
    {
      "network": "ethereum",
      "rpc_url": "https://ethereum-rpc.publicnode.com",
      "poll_interval": "12s",
      "aave": {
        "pool": "0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2",
        "data_provider": "0x...",
        "oracle": "0x...",
        "wallets": ["0x..."]
      },
      "morpho": {
        "address": "0xBBBBBbbBBb9cC5e90e3b3Af64bdAF62C37EEFFCb",
        "deploy_block": 19336000,
        "parallelism": 100,
        "scan_batch_size": 10000,
        "wallets": ["0x..."]
      }
    }
  ]
}
```

У каждого протокола свои кошельки и параметры. У Morpho — `deploy_block` (свой для каждой сети), `parallelism` (сколько горутин проверяют рынки, по умолчанию 100) и `scan_batch_size` (сколько блоков за один запрос при первичном сканировании `CreateMarket`, по умолчанию 10000).

## Таблица `positions`

Колонки: `protocol` (aave-v3 / morpho-blue), `network` (ethereum / arbitrum), `wallet_address`, `market_id`, `collateral_token`, `debt_token`, `position_size`, `token_price` (цена от оракула, USD), `health_factor`, `block_number`, `timestamp`, `created_at`.

## Допущения

- Supply-only позиции не сохраняются - задание явно про долговые позиции, а не про все, если судить по тому, что нужен debt token.
- Сканируются все блоки с текущего на момент запуска.
- В качестве хранилища использую Postgres, но если контекста сервиса скорее про скармливание позиций другим сервисам, то можно commit log Kafka или обычная очередь RabbitMQ, достаточно реализовать интерфейс `positionSaver`.
- Использую логи там, где нужны по-хорошему метрики Прометея, но в задаче сказано логи, значит только логи.
- Не учитываю редчайший кейс реорганизации блокчейна, когда блоки откатываются назад, в этом случае вся система, включая потенциальных потребителей должна его учитывать.
- Если оракул не отвечает по цене, то позиция пропускается после нескольких ретраев.

## RPC

Сейчас публичные эндпоинты — `ethereum-rpc.publicnode.com` и `arb1.arbitrum.io/rpc`.
Для прода явно нужны выделенные: публичные дают мало соединений, режут по лимитам и иногда отваливаются.

## Тесты

```bash
go test ./...
```

## Как добавить протокол или сеть

Протокол - структура, реализующая интерфейс парсинга позиций
```go
type positionsParser interface {
ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]Position, error)
}
```
плюс блок в `jobs.json` плюс поля в `config.go` плюс пробросить в ините в `app.go`.

Новая EVM-сеть - просто ещё один джоб в `jobs.json` со своими адресами и названием сети, которое будет писаться в `network` колонку.
Код не нужно менять, при интеграции Arbitrum проблем не было, кроме ограничений публичного эндпоинта.
