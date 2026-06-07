# VLW - VRC Log Watcher

VRChat のログファイルをリアルタイムに監視し、特定の行を検知したときに通知やアクションを実行する Windows デスクトップアプリです。

[Wails v2](https://wails.io/) (Go + Svelte/TypeScript) で構築されています。

## できること

VRChat のログを読み込み、新しく追記された行を正規表現で評価します。  
マッチした行に対して、設定ごとに以下のアクションを実行できます。

| Type | 動作 |
|------|------|
| `LogOnly` | アプリ内のログ表示のみ |
| `WebRequest` | 指定した URL へ HTTP POST |
| `SendXSOverlay` | XSOverlay へ UDP で通知を送信 (ポート 42069) |
| `SendDiscordWebHook` | Discord Webhook へメッセージ送信 (画像添付対応) |
| `OutputTextFile` | マッチ結果をテキストファイルへ出力 |
| `Disable` | 何もしない |

URL や送信内容には `{time}` `{title}` `{description}` `{matched_text}` `{log_file}` といったテンプレート変数を埋め込めます。

「ワールド入室」「ユーザーの入退室」「スクリーンショット撮影」などを検知する設定がプリセット (Core Settings) として最初から用意されており、自分で正規表現を組んで独自の検知ルールを追加することもできます。

## 設定ファイル

`setting.json` は実行ファイルと同じディレクトリに保存されます。初回起動時に自動生成され、VRChat のログ出力先を自動検出して書き込まれます。

## 開発情報

[mise](https://mise.jdx.dev/) でツールバージョンを管理しています ([.mise.toml](.mise.toml))。

- Go 1.22+
- Node.js 24+
- pnpm 9+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

## コマンド

```bash
# 開発モードで起動
wails dev

# 本番ビルド
wails build

# バインディングを再生成
wails generate module
```

mise のタスクからも同様の操作が可能です (`mise run dev` / `mise run build` など)。

## ライセンス

[LICENSE](LICENSE) 
