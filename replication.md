# Replication: Current Flow and Testing

## How replication works right now

1. Server role selection on startup.
- `cmd/server/main.go` reads args.
- Default role is `master`.
- If started with `--replicaof <host> <port>`, role becomes `slave`.

2. Replica initiates connection to master.
- In slave mode, `go helper.ConnectToMaster(st)` is started.
- Replica dials `masterHost:masterPort`.

3. Replica performs handshake on the same TCP connection.
- Sends `PING`.
- Sends `REPLCONF listening-port <replica-port>`.
- Sends `REPLCONF capa psync2`.
- Sends `PSYNC ? -1`.

4. Master handles `PSYNC` and registers replica connection.
- On `PSYNC`, master sends `+FULLRESYNC <replid> <offset>`.
- Sends an RDB bulk payload (currently `emptyRDB`).
- Appends that same `conn` into `st.Replicas`.

5. Replica reads snapshot then enters stream loop.
- Reads FULLRESYNC line.
- Reads RDB bulk header and exact binary payload.
- Enters loop reading RESP arrays from the master connection.
- Applies incoming commands via `commands.ApplyReplicaCommand(st, parts)`.

6. Master propagates commands to registered replicas.
- After command execution in normal request flow, master calls `helper.PropagateToReplicas(st, parts)`.
- `PropagateToReplicas` serializes the command to RESP and writes to each `st.Replicas` socket.

7. Effective replicated command support today.
- Replica apply path currently handles only `SET`.
- Other propagated commands are ignored by `ApplyReplicaCommand`.

## Important current limitations

- `master_repl_offset` is not incremented during propagation.
- Dead/disconnected replica sockets are not removed when writes fail.
- Replica command apply is limited (currently only `SET`).

## Manual test steps

## 1) Build

```bash
go build -o go_redis_server ./cmd/server
```

## 2) Start master (Terminal A)

```bash
./go_redis_server --port 6380
```

## 3) Start replica (Terminal B)

```bash
./go_redis_server --port 6381 --replicaof 127.0.0.1 6380
```

## 4) Write on master (Terminal C)

```bash
redis-cli -p 6380 SET mykey hello
redis-cli -p 6380 GET mykey
```

Expected:
- Master returns `OK` for `SET` and `hello` for `GET`.

## 5) Read on replica

```bash
redis-cli -p 6381 GET mykey
```

Expected:
- Replica returns `hello`.

## 6) Verify replication role reporting

```bash
redis-cli -p 6380 INFO replication
redis-cli -p 6381 INFO replication
```

Expected:
- Master info contains `role:master`.
- Replica info contains `role:slave`.

## 7) Show current command-coverage gap (`INCR` example)

```bash
redis-cli -p 6380 SET c 1
redis-cli -p 6380 INCR c
redis-cli -p 6380 GET c
redis-cli -p 6381 GET c
```

Expected right now:
- Master `GET c` is `2`.
- Replica `GET c` may still be `1` because replica apply path currently handles only `SET`.

## 8) Optional raw RESP test (if `redis-cli` is unavailable)

```bash
printf '*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$5\r\nhello\r\n' | nc 127.0.0.1 6380
printf '*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n' | nc 127.0.0.1 6381
```

Expected:
- Replica returns `hello` for `mykey`.
