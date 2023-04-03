# Flare P-chain Attestation Suite

Project consists of two projects

* P-chain indexer (indexer)
* Attestation client (services)

## P-chain Indexer

Indexer periodically reads blocks from Avalanche-Go (Flare) node with
enabled indexing (parameter `--index-enabled` set to true) from `/ext/index/P/block` route and writes transactions and their UTXO inputs and outputs to a MySQL database.

Executable can be compiled from `indexer/main/main.go`.

Configuration is read from `toml` file `config.toml`. Some configuration
parameters can also be configured from environment variables. See example below.

```toml
[db]
host = "localhost"  # MySql db address, or env variable DB_HOST
port = 3306         # MySql db port, env DB_PORT
database = "flare_indexer"    # database name, env DB_DATABASE
username = "indexeruser"      # db username, env DB_USERNAME
password = "P.a.s.s.W.O.R.D"  # db password, env DB_PASSWORD
log_queries = false  # Log db queries (for debugging)

[logger]
level = "INFO"      # valid values are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL (as in zap logger)
file = "./logs/flare-indexer.log"  # logger file
max_file_size = 10  # max file size before rotating
console = true      # also log to console

[metrics]
prometheus_address = "localhost:2112"  # expose indexer metrics to this address (empty value does not expose this endpoint)

[chain]
node_url = "http://localhost:9650/"  # node indexer address
address_hrp = "localflare"  # HRP (human readable part) of chain -- used to properly encode/decode addresses
chain_id = 162  # chain id

[p_chain_indexer]
enabled = true         # enable p-chain indexing
timeout_millis = 1000  # call avalanche p-chain indexer every ... ms
batch_size = 10        # batch size to fetch from the node (max ????)
start_index = 0        # start indexing at this block height

[uptime_cronjob]
enabled = false       # enable uptime monitoring cronjob
timeout_seconds = 10  # call uptime service on avalanche node evey
```

## Attestation client services

Implements the following services

* `/query/`
* `/query/prepare`
* `/query/integrity`
* `/query/prepareAttestation`

according to attestation specification.

Configuration is read from `toml` file `config.toml`.
Settings for `[db]`, `[logger]` are the same as for the indexer above.
Specific settings are

```toml
[chain]
address_hrp = "localflare"  # HRP (human readable part) of chain -- used to properly encode/decode addresses
chain_id = 162  # chain id

[services]
address = "localhost:8000"  # address and port to run the server at
```
