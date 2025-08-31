# 🥸 Stache

A simple in-memory cache written in Go — started as a hobby project to learn Go.

## Features
- **In-memory key/value store** with optional TTL expiry
- **MIME Support for**: `text/plain` and `application/json`
- **Thread-safe**: built with sync.RWMutex
- **Introspection**: list entries with metadata (size, content-type, expiry)
- **Library errors**: typed sentinel errors (ErrNotFound, ErrIncorrectType)

## API
-	**Protobuf definitions** in api/stache/v1
-	**ConnectRPC server**: serves
    - gRPC
    - gRPC-Web
    - Connect protocol (HTTP/1.1 or HTTP/2)
-	**Self-documenting**: reflection enabled for grpcurl

## Daemon
- stached runs the cache server
- Supports h2c (HTTP/2 cleartext) for local dev
- Graceful shutdown with signal handling
- Ready to run behind TLS

## CLI
- Built-in CLI client (cmd/stache) for quick interaction:

```bash
stache -set name -v "dababy" -t text/plain -l 60
stache -get name
stache -list
```

- Uses generated Connect client stubs
- Pretty-prints JSON responses and tabular listings

## TBD
- On-disk persistence
- Additional eviction policies
- Metrics
- Authentication etc.

*<small>Still experimental and not production-ready! Pls do not use in anything that's real</small>*
