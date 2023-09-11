# Flare P-chain Attestation Suite

This code implements two projects (compiled into two executables)

* P-chain clients
* Attestation client (services)

## P-chain Clients

The executable can be built with `go build indexer/main/indexer.go`.
This executable consists of several clients/cronjobs which can be enabled/disabled in the configuration file.

* P-chain indexer
* Uptime monitoring cronjob
* Voting client
* Mirroring client

### P-chain indexer

The P-chain indexer periodically reads blocks from an Avalanche-Go (Flare) node with
enabled indexing (parameter `--index-enabled` set to true) from `/ext/index/P/block` route and writes transactions and their UTXO inputs and outputs to a MySQL database.

### Uptime monitoring cronjob

The uptime monitoring cronjob periodically calls the `platform.getCurrentValidators` P-chain API route and writes all current validator node IDs thogether with "connected" flag to a MySQL database.

### Voting client

The voting client fetches all validators or delegators starting in a particular epoch from the MySQL database, creates a Merkle tree of their data hashes, and sends a vote transaction (epoch and Merkle tree root) to the voting contract.
This is done for all epoch not already processes or voted for.

### Mirroring client

Sends the data about validators in a particuler epoch to the mirror contract.

### Configuration

The configuration is read from `toml` file. Some configuration
parameters can also be configured using environment variables. See the list below.

Config file can be specified using the command line parameter `--config`, e.g., `./indexer --config config.local.toml`. The default config file name is `config.toml`.

Below is the list of configuration parameters for all clients. Clients that are not enabled can be omitted from the config file.

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
max_file_size = 10  # max file size before rotating, in MB
console = true      # also log to console

[metrics]
prometheus_address = "localhost:2112"  # expose indexer metrics to this address (empty value does not expose this endpoint)

[chain]
node_url = "http://localhost:9650/"  # node indexer address
address_hrp = "localflare"  # HRP (human readable part) of chain -- used to properly encode/decode addresses
chain_id = 162  # chain id
eth_rpc_url = "http://localhost:9650/ext/C/rpc"  # Ethereum RPC URL
api_key = ""    # API key (in case the node is protected by API key), adds ?x-apikey=... to all requests if not empty
private_key = "0x00000000000000"  # private key of the account (for voting and mirroring clients), in hex

[p_chain_indexer]
enabled = true         # enable p-chain indexing
timeout = "1000ms"     # call avalanche p-chain indexer every ...
batch_size = 10        # batch size to fetch from the node
start_index = 0        # start indexing at this block height

[uptime_cronjob]
enabled = false         # enable uptime monitoring cronjob
timeout = "10s"         # call uptime service on every ...
enable_voting = true    # enable voting for connected validators
start = "2021-08-01T00:00:00Z"  # start of the uptime voting epoch, supports also unix timestamp as a number
period = "90s"          # length of the epoch
uptime_threshold = 0.8  # minimum uptime ratio in the epoch for a validator to be considered connected

# epochs for mirroring and voting clients
[epochs]
start = "2021-08-01T00:00:00Z"  # start of the voting epochs, supports also unix timestamp as a number
period = "90s"          # length of the epoch

[voting_cronjob]
enabled = false          # enable voting client
timeout = "10s"          # check for new epochs every ...
contract_address = "0xf956df3800379fdFA31D0A45FDD5001D02F4109c"    # voting contract address

[mirroring_cronjob]
enabled = false       # enable mirroring client
timeout = "10s"       # check for new epochs every ... seconds
contract_address = "0xE64Df6a7e4f4c277C5299f0FE12D7BbB8A207175"    # mirror contract address
```

### Running tests

Tests for indexer expect a MySQL database running. Please run `docker-compose up` in the `indexer/resources/test` directory.

Additionally, tests for voting, mirroring and uptime clients expect a Hardhat instance from <https://gitlab.com/flarenetwork/flare-smart-contracts/-/tree/staking-tests> running. You start it by running
`yarn staking_test` (following `yarn` and `yarn c` commands).

Then run `go test ./...` in the `indexer` directory.

## Attestation client services (possible future use)

The following services are implemented, according to the attestation specification:

* `/query`
* `/query/prepare`
* `/query/integrity`
* `/query/prepareAttestation`

The executable can be built with with `go build services/main/services.go`.

The configuration is read from `toml` file.
The settings for `[db]`, `[logger]` are the same as for the indexer above.
Specific settings are listed below.

Config file can be specified using the command line parameter `--config`, e.g., `./services --config config.local.toml`. The default config file name is `config.toml`.

```toml
[chain]
address_hrp = "localflare"  # HRP (human readable part) of chain -- used to properly encode/decode addresses
chain_id = 162  # chain id

[services]
address = "localhost:8000"  # address and port to run the server at
```
