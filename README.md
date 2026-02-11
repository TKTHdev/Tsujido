# tsujido
[English](README.md) | [日本語](README.ja.md)
`tsujido` is a lightweight Go framework designed for building and benchmarking distributed system protocols (such as Chain Replication, Raft, PBFT). It provides the common infrastructure—networking, serialization, state machine, and benchmarking tools—allowing developers to focus on implementing the consensus or replication logic.

## Features

- **Networking**: Robust TCP-based `Server` and `Client` implementations.
- **Wire Protocol**: Efficient, custom binary serialization for operations (GET, SET, DELETE).
- **State Machine**: Built-in in-memory Key-Value Store (`KVStore`) supporting standard operations.
- **Benchmarking**: Integrated `YCSBRunner` to simulate YCSB workloads (A, B, C) and measure throughput/latency.

## Installation

```bash
go get github.com/TKTHdev/tsujido
```

## Usage

To use `tsujido`, you typically implement the `RequestHandler` interface with your specific protocol logic.

### 1. Implement a Request Handler

```go
package main

import (
	"context"
	"github.com/TKTHdev/tsujido"
)

type MyProtocol struct {
	sm tsujido.StateMachine
}

func (p *MyProtocol) HandleRequest(ctx context.Context, op tsujido.Operation) (tsujido.Result, error) {
	// Implement your consensus/replication logic here.
	// For example, replicate to other nodes before applying.
	
	// Apply to state machine locally
	return p.sm.Apply(op), nil
}
```

### 2. Start the Server

```go
func main() {
	sm := tsujido.NewKVStore()
	handler := &MyProtocol{sm: sm}
	
	server := tsujido.NewServer(":8080", handler)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
```

### 3. Run a Benchmark (YCSB)

You can easily build a client to benchmark your implementation using the built-in YCSB runner.

```go
func main() {
	client, _ := tsujido.NewTCPClient("localhost:8080", "localhost:8080")
	
	config := tsujido.YCSBConfig{
		Workers:  10,
		Workload: 50, // 50% writes (YCSB-A)
		Duration: 10 * time.Second,
		Protocol: "my-protocol",
	}
	
	runner := tsujido.NewYCSBRunner(client, config)
	runner.Run()
}
```

## Structure

- **`client.go`**: TCP client implementation.
- **`server.go`**: TCP server listener and connection handler.
- **`protocol.go`**: Wire format definitions and serialization helpers.
- **`statemachine.go`**: `KVStore` implementation and `StateMachine` interface.
- **`ycsb.go`**: YCSB load generator and statistics collector.
