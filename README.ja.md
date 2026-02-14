# tsujido
[English](README.md) | [日本語](README.ja.md)

`tsujido` は、分散システム向けのステートマシンライブラリです。チャネルベースの逐次書き込みによる KV ストアと、`Endpoint` によるコンセンサス層との統合インターフェースを提供します。

## 特徴

- **Query / Handle**: 読み取りは書き込みチャネルをバイパス (`RLock`)、書き込みは単一 goroutine で逐次実行。
- **コンセンサス統合**: `Endpoint` の `ProposeCh` / `CommitCh` を通じて外部のコンセンサス層 (Raft, Paxos 等) を接続可能。

## インストール

```bash
go get github.com/TKTHdev/tsujido
```

## 使い方

```go
ep := tsujido.NewEndpoint()
defer ep.Stop()

// ProposeCh と CommitCh の間にコンセンサス層を接続する。
go func() {
    for req := range ep.ProposeCh {
        // ここで Raft 等の合意処理を行う。
        ep.CommitCh <- req
    }
}()

// 書き込み
res := ep.Handle("client-1", tsujido.Operation{Key: "x", Value: "hello"})
// → Result{Success: true, Value: "OK"}

// 読み取り
res = ep.Query("x")
// → Result{Success: true, Value: "hello"}
```

### フロー

```
Handle()  →  ProposeCh  →  [コンセンサス層]  →  CommitCh  →  状態適用
  ↑                                                            |
  └─────────────── req.done 経由で Result が返る ───────────────┘
```

## API

| メソッド / フィールド | 説明 |
|---|---|
| `NewEndpoint()` | Endpoint を生成（内部 goroutine を起動） |
| `Handle(clientID, op)` | 書き込みリクエスト（合意 + 適用完了までブロック） |
| `Query(key)` | 値の読み取り（ノンブロッキング、`RLock` 使用） |
| `Stop()` | ステートマシンを停止 |
| `ProposeCh` | 提案されたリクエストを読み取るチャネル |
| `CommitCh` | 合意済みリクエストを投入するチャネル |

## 構成

- **`types.go`** — `Operation`, `Result`, `Request`
- **`statemachine.go`** — 内部ステートマシン（チャネルベースの逐次書き込み KV ストア）
- **`endpoint.go`** — `Endpoint`（公開 API とコンセンサス層との接続）
