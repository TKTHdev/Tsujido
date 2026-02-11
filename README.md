# tsujido
[English](README.md) | [日本語](README.ja.md)

`tsujido` is an idempotent state machine for distributed systems in Go.

## Features

- **Query / Submit**: Reads bypass the write channel (`RLock`), writes are serialized through a single goroutine.
- **Idempotency**: Duplicate writes (same `ClientID` + `Seq`) return the cached result without re-execution.

## Installation

```bash
go get github.com/TKTHdev/tsujido
```

## Usage

```go
sm := tsujido.NewStateMachine()
defer sm.Stop()

// Write
sm.Submit(
    tsujido.RequestID{ClientID: "10.0.0.1:40000", Seq: 1},
    tsujido.Operation{Key: "x", Value: "hello"},
) // {Success: true, Value: "OK"}

// Read
sm.Query("x") // {Success: true, Value: "hello"}

// Replay Seq=1 → cached, not re-executed
sm.Submit(
    tsujido.RequestID{ClientID: "10.0.0.1:40000", Seq: 1},
    tsujido.Operation{Key: "x", Value: "overwrite"},
) // {Success: true, Value: "OK"} (cached, "x" is still "hello")
```

## Structure

- **`types.go`** — `Operation`, `RequestID`, `Result`
- **`statemachine.go`** — `StateMachine` (idempotent KV store with channel-based serial writes)
