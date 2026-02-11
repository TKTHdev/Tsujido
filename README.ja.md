# tsujido
[English](README.md) | [日本語](README.ja.md)

`tsujido` は、分散システム向けの冪等性保証付きステートマシンです。

## 特徴

- **Query / Submit**: 読み取りは書き込みチャネルをバイパス (`RLock`)、書き込みは単一 goroutine で逐次実行。
- **冪等性**: 同じ `ClientID` + `Seq` の書き込みはキャッシュされた結果を返し、再実行されない。

## インストール

```bash
go get github.com/TKTHdev/tsujido
```

## 使い方

```go
sm := tsujido.NewStateMachine()
defer sm.Stop()

// 書き込み
sm.Submit(
    tsujido.RequestID{ClientID: "10.0.0.1:40000", Seq: 1},
    tsujido.Operation{Key: "x", Value: "hello"},
) // {Success: true, Value: "OK"}

// 読み取り
sm.Query("x") // {Success: true, Value: "hello"}

// Seq=1 を再送 → キャッシュが返り、再実行されない
sm.Submit(
    tsujido.RequestID{ClientID: "10.0.0.1:40000", Seq: 1},
    tsujido.Operation{Key: "x", Value: "overwrite"},
) // {Success: true, Value: "OK"} (キャッシュ、"x" は "hello" のまま)
```

## 構成

- **`types.go`** — `Operation`, `RequestID`, `Result`
- **`statemachine.go`** — `StateMachine`（チャネルベースの逐次書き込み + 冪等性付き KV ストア）
