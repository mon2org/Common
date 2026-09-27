FROM debian:bookworm-slim@sha256:3783cc01769c7b2b1b83a5c5ad96c815348e28ed7da68e2e3687004faa906251

# XDG Base Directory 環境変数の定義
ENV HOME="/root"
ENV XDG_DATA_HOME="${HOME}/.local/share"
ENV XDG_CONFIG_HOME="${HOME}/.config"
ENV XDG_CACHE_HOME="${HOME}/.cache"
ENV XDG_STATE_HOME="${HOME}/.local/state"

# PATH の設定
ENV PATH="${HOME}/.local/bin:${XDG_DATA_HOME}/mise/shims:$PATH"

# 1. 基本パッケージと zsh のインストール
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    zsh \
    && rm -rf /var/lib/apt/lists/*

# 2. インストールする mise のバージョンを指定して公式スクリプトを実行
ARG MISE_VERSION=v2026.9.15
RUN MISE_VERSION=${MISE_VERSION} curl https://mise.run | sh

# 3. ホスト側の config ディレクトリが存在する場合に .config へコピー
# ※ config ディレクトリが空（.gitkeepのみ等）の場合は中身がコピーされないため安全です
COPY config/ ${XDG_CONFIG_HOME}/

# 4. bash と zsh の両方で mise が有効化されるように設定
RUN echo 'eval "$(mise activate bash)"' >> "${HOME}/.bashrc" \
    && echo 'eval "$(mise activate zsh)"' >> "${HOME}/.zshrc"

# 5. 動作確認
RUN mise --version