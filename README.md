# tsujido
[English](README.md) | [日本語](README.ja.md)

`tsujido` is a state machine library for distributed systems in Go. It provides a KV store with channel-based serial writes and a consensus integration layer via `Endpoint`.

## Features

- **Query / Handle**: Reads bypass the write channel (`RLock`), writes are serialized through a single goroutine.
- **Consensus integration**: `Endpoint` exposes `ProposeCh` / `CommitCh` to plug in an external consensus layer (Raft, Paxos, etc.).

## Installation

```bash
go get github.com/TKTHdev/tsujido
```

## Usage

```go
ep := tsujido.NewEndpoint()
defer ep.Stop()

// Plug in your consensus layer between ProposeCh and CommitCh.
go func() {
    for req := range ep.ProposeCh {
        // Run consensus (e.g. Raft) here, then commit.
        ep.CommitCh <- req
    }
}()

// Write
res := ep.Handle("client-1", tsujido.Operation{Key: "x", Value: "hello"})
// → Result{Success: true, Value: "OK"}

// Read
res = ep.Query("x")
// → Result{Success: true, Value: "hello"}
```

### Flow

```
Handle()  →  ProposeCh  →  [Consensus Layer]  →  CommitCh  →  state applied
  ↑                                                                 |
  └─────────────── Result returned via req.done ────────────────────┘
```

## API

| Method / Field | Description |
|---|---|
| `NewEndpoint()` | Create an Endpoint (starts internal goroutines) |
| `Handle(clientID, op)` | Submit a write request (blocks until consensus + apply) |
| `Query(key)` | Read a value (non-blocking, uses `RLock`) |
| `Stop()` | Shut down the state machine |
| `ProposeCh` | Channel to read proposed requests from |
| `CommitCh` | Channel to send committed requests to |

## Structure

- **`types.go`** — `Operation`, `Result`, `Request`
- **`statemachine.go`** — internal state machine (KV store with channel-based serial writes)
- **`endpoint.go`** — `Endpoint` (public API and consensus integration)
