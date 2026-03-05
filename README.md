# devcontainer-sandbox

汎用・使い捨て Dev Container。最小ベースイメージ + 公式 Dev Container Features で構成を選べる。

## コンセプト

- **最小ベース**: `mcr.microsoft.com/devcontainers/base:ubuntu-24.04` をベースに、必要なものだけ追加
- **Features で選択式**: 言語ランタイムやツールはコメントを外すだけで追加
- **ファイアウォール付き**: 許可ドメインリストに基づいて外部通信を制限（オプション）
- **使い捨て可能**: SSH マウント不要、ボリューム名に `devcontainerId` を使用して独立性を確保

## ファイル構成

```
.
├── .devcontainer/
│   ├── Dockerfile              # 最小ベースイメージ + ファイアウォール設定
│   ├── devcontainer.json       # Dev Container 設定（Features・マウント・拡張機能）
│   ├── init-firewall.sh        # ファイアウォール初期化スクリプト
│   └── allowed-domains.conf    # 許可ドメインリスト
└── README.md
```

## 使い方

### 言語・ツールの選択

`devcontainer.json` の `features` セクションのコメントを外す。

```jsonc
"features": {
  // --- Languages ---
  "ghcr.io/devcontainers/features/node:1": { "version": "22" },  // デフォルト有効
  // "ghcr.io/devcontainers/features/python:1": { "version": "3.12" },
  // "ghcr.io/devcontainers/features/rust:1": { "profile": "minimal" },
  // "ghcr.io/devcontainers/features/go:1": { "version": "latest" },
  // "ghcr.io/devcontainers/features/java:1": { "version": "21" },
  // "ghcr.io/devcontainers/features/ruby:1": { "version": "latest" },

  // --- Tools ---
  "ghcr.io/devcontainers/features/github-cli:1": {},              // デフォルト有効
  // "ghcr.io/devcontainers/features/docker-in-docker:2": {},
  // "ghcr.io/devcontainers/features/kubectl-helm-minikube:1": {},
  // "ghcr.io/devcontainers/features/terraform:1": {}
}
```

不要な言語は行ごと削除するか、コメントアウトすればビルド時に含まれない。

### AI ツール

コンテナ作成後に自動インストールされる（`postCreateCommand`）:

- `claude` (@anthropic-ai/claude-code)
- `codex` (@openai/codex)
- `pnpm`

### ファイアウォール

起動時に `init-firewall.sh` が自動実行され、`allowed-domains.conf` に記載されたドメインと GitHub IP 範囲のみ通信を許可する。

許可ドメインを追加したい場合は `allowed-domains.conf` に1行ずつ追加:

```
# allowed-domains.conf
registry.npmjs.org
api.anthropic.com
your-domain.example.com
```

ファイアウォールを無効化したい場合は `devcontainer.json` の `postStartCommand` を変更:

```jsonc
"postStartCommand": "true"
```

### ボリューム

| ボリューム名 | マウント先 | 用途 |
|---|---|---|
| `sandbox-bashhistory-{id}` | `/commandhistory` | コマンド履歴（コンテナ固有） |
| `sandbox-claude-config-{id}` | `/home/dev/.claude` | Claude 設定（コンテナ固有） |
| `sandbox-projects` | `/home/dev/projects` | プロジェクト（コンテナ間共有） |

## 検証

```bash
# ファイアウォールが機能しているか確認
curl https://example.com          # 失敗するはず
curl https://api.github.com/zen   # 成功するはず

# AI ツールが使えるか確認
claude --version
codex --version
```
