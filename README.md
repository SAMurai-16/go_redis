# CacheFlow


A small Redis-compatible server written in Go. It speaks RESP over TCP, stores data in memory, and implements a focused subset of Redis commands across strings, lists, streams, pub/sub, transactions, persistence loading, and basic replication.

This project is useful for learning how Redis works internally: request parsing, command dispatch, key expiry, stream IDs, replication handshakes, and RESP encoding are all implemented directly in the codebase.

## Features

- RESP array parsing for Redis CLI compatible requests
- TCP server with concurrent client handling
- String commands: `PING`, `ECHO`, `SET`, `GET`, `INCR`
- Key commands: `TYPE`, `KEYS`
- List commands: `LPUSH`, `RPUSH`, `LRANGE`, `LLEN`, `LPOP`
- Stream commands: `XADD`, `XRANGE`, `XREAD`
- Pub/sub commands: `SUBSCRIBE`, `UNSUBSCRIBE`, `PUBLISH`
- Transactions: `MULTI`, `EXEC`, `DISCARD`
- Replication commands: `INFO`, `REPLCONF`, `PSYNC`, `WAIT`
- RDB loading for string keys from a Redis RDB file
- Config lookup for `dir` and `dbfilename`

## Requirements

- Go 1.25.6 or newer, matching `go.mod`
- Optional: `redis-cli` for manual testing

## Build

```bash
go build -o go_redis_server ./cmd/server
```

## Run

Start a standalone server:

```bash
./go_redis_server --port 6380
```

The server listens on port `6380` by default if `--port` is omitted.

Connect with Redis CLI:

```bash
redis-cli -p 6380
```

Example session:

```bash
redis-cli -p 6380 SET greeting hello
redis-cli -p 6380 GET greeting
redis-cli -p 6380 INCR counter
redis-cli -p 6380 TYPE greeting
```

## Startup Options

```bash
./go_redis_server [--port <port>] [--replicaof <host> <port>] [--dir <path>] [--dbfilename <file>]
```

Options:

- `--port <port>`: TCP port to listen on. Defaults to `6380`.
- `--replicaof <host> <port>`: start this server as a replica of the given master.
- `--dir <path>`: directory used when loading an RDB file. Defaults to `.`.
- `--dbfilename <file>`: RDB filename to load. Defaults to `dump.rdb`.

## Persistence

On startup, the server attempts to load an RDB file from:

```text
<dir>/<dbfilename>
```

For example:

```bash
./go_redis_server --port 6380 --dir /tmp/redis-data --dbfilename dump.rdb
```

The current RDB loader supports Redis RDB string values, including keys with second or millisecond expiry metadata.

## Replication

Start a master:

```bash
./go_redis_server --port 6380
```

Start a replica:

```bash
./go_redis_server --port 6381 --replicaof 127.0.0.1 6380
```

Try a replicated write:

```bash
redis-cli -p 6380 SET mykey hello
redis-cli -p 6381 GET mykey
```

Check roles:

```bash
redis-cli -p 6380 INFO replication
redis-cli -p 6381 INFO replication
```

Replication is intentionally minimal. The server performs the `PING`, `REPLCONF`, and `PSYNC` handshake, sends an empty RDB snapshot for full sync, and streams propagated write commands from master to replica. The replica apply path currently handles `SET`; other propagated write commands are still limited.

See [replication.md](replication.md) for a more detailed walkthrough and manual test plan.

## Project Structure

```text
cmd/server/       TCP server entrypoint and connection loop
commands/         Redis command handlers
engine/           Command dispatcher
helper/           Replication connection and propagation helpers
resp/             RESP parser
store/            In-memory data store, streams, RDB loading, counters
replication.md    Replication notes and manual test steps
```

## Development

Run all Go package checks:

```bash
go test ./...
```

Format code:

```bash
go fmt ./...
```

## Notes and Limitations

- This is a learning implementation, not a production Redis replacement.
- Data is stored in memory; RDB support currently loads data on startup but does not save snapshots.
- Replication support is partial and primarily demonstrates the full-sync handshake and command propagation.
- Pub/sub subscriptions are connection-local and managed in memory.
- Stream reads are non-blocking.
