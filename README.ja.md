# tsujido
[English](README.md) | [日本語](README.ja.md)
`tsujido` は、分散システムプロトコル（Chain Replication, Raft, PBFTなど）を構築・ベンチマークするための軽量な Go フレームワークです。ネットワーク通信、シリアライゼーション、ステートマシン、ベンチマークツールといった共通インフラを提供し、開発者がコンセンサスやレプリケーションのロジックの実装に集中できるように設計されています。

## 特徴

- **ネットワーキング**: 堅牢な TCP ベースの `Server` および `Client` 実装。
- **ワイヤープロトコル**: 操作（GET, SET, DELETE）のための効率的なカスタムバイナリシリアライゼーション。
- **ステートマシン**: 標準的な操作をサポートする内蔵インメモリ Key-Value ストア (`KVStore`)。
- **ベンチマーク**: YCSB ワークロード (A, B, C) をシミュレートし、スループットとレイテンシを計測する `YCSBRunner` を統合。

## インストール

```bash
go get github.com/TKTHdev/tsujido
```

## 使い方

`tsujido` を使用するには、通常 `RequestHandler` インターフェースを実装し、独自のプロトコルロジックを記述します。

### 1. Request Handler の実装

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
	// ここにコンセンサスやレプリケーションのロジックを実装します。
	// 例: 適用する前に他のノードへレプリケーションを行うなど。
	
	// ステートマシンに適用
	return p.sm.Apply(op), nil
}
```

### 2. サーバーの起動

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

### 3. ベンチマークの実行 (YCSB)

組み込みの YCSB ランナーを使用して、実装のベンチマークを行うクライアントを簡単に作成できます。

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

## 構成

- **`client.go`**: TCP クライアントの実装。
- **`server.go`**: TCP サーバーのリスナーおよび接続ハンドラー。
- **`protocol.go`**: ワイヤーフォーマットの定義とシリアライゼーションヘルパー。
- **`statemachine.go`**: `KVStore` の実装と `StateMachine` インターフェース。
- **`ycsb.go`**: YCSB 負荷生成ツールと統計収集。
