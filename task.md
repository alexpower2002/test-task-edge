Тестовое задание: Backend Developer

Реализовать сервис для парсинга lending-позиций по заданному списку кошельков в Ethereum mainnet.

Обязательный протокол: Aave v3.
Дополнительно один на выбор: Morpho, Euler или Fluid.

Архитектура должна позволять относительно просто добавить новый протокол или другую EVM-сеть.

На каждом новом блоке сервис должен обновлять позиции для заданного списка wallet-адресов.

Для каждой позиции получать: protocol, wallet address, market/pool/vault identifier, collateral token, debt token, position size, token price, health ratio / health factor, block number, timestamp.

Цена токена должна браться только из on-chain источников: protocol oracle, Chainlink, pool pricing и т.д. Использование CoinGecko и аналогичных API запрещено.

Требования:
• запуск через docker-compose;
• тесты;
• логи;
• README.

Разрешено использование любых ИИ-инструментов.

Стек: Go.
