FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

ARG TZ=Asia/Tokyo
ENV TZ="$TZ"
ARG USERNAME=dev
ARG DELTA_VERSION=0.18.2

# Create dev user
RUN useradd -m -s /bin/zsh -G sudo ${USERNAME} && \
  echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME}

# Minimal extra tools
RUN apt-get update && apt-get install -y --no-install-recommends \
  fzf man-db nano vim \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# git-delta
RUN ARCH=$(uname -m) && \
  if [ "$ARCH" = "x86_64" ]; then DELTA_ARCH="x86_64-unknown-linux-musl"; else DELTA_ARCH="aarch64-unknown-linux-gnu"; fi && \
  curl -LO "https://github.com/dandavison/delta/releases/download/${DELTA_VERSION}/delta-${DELTA_VERSION}-$DELTA_ARCH.tar.gz" && \
  tar -xzf "delta-${DELTA_VERSION}-$DELTA_ARCH.tar.gz" --strip-components=1 -C /usr/local/bin "delta-${DELTA_VERSION}-$DELTA_ARCH/delta" && \
  rm "delta-${DELTA_VERSION}-$DELTA_ARCH.tar.gz"

# Persist bash history
RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  && mkdir /commandhistory \
  && touch /commandhistory/.bash_history \
  && chown -R ${USERNAME} /commandhistory

ARG PNPM_VERSION=10.30.3
ARG CLAUDE_CODE_VERSION=2.1.69
ARG CODEX_VERSION=0.106.0

# Config directories
RUN mkdir -p /home/${USERNAME}/.claude && \
  chown -R ${USERNAME}:${USERNAME} /home/${USERNAME}/.claude

ENV DEVCONTAINER=true
ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano

# Firewall (optional: only functional when init-firewall.sh is present)
COPY init-firewall.sh firewall-status.sh /usr/local/bin/
COPY allowed-domains.conf /etc/allowed-domains.conf
COPY postCreate.sh /usr/local/bin/postCreate.sh
RUN chmod +x /usr/local/bin/init-firewall.sh /usr/local/bin/firewall-status.sh /usr/local/bin/postCreate.sh && \
  echo "${USERNAME} ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" > /etc/sudoers.d/${USERNAME}-firewall && \
  chmod 0440 /etc/sudoers.d/${USERNAME}-firewall

USER ${USERNAME}

# Install AI tools at build time
ENV PNPM_HOME="/home/${USERNAME}/.local/share/pnpm"
ENV PATH="${PNPM_HOME}:${PATH}"
RUN npm install -g "pnpm@${PNPM_VERSION}" && \
  mkdir -p "${PNPM_HOME}" && \
  grep -qF 'PNPM_HOME' ~/.zshrc || printf 'export PNPM_HOME="${HOME}/.local/share/pnpm"\nexport PATH="${PNPM_HOME}:${PATH}"\n' >> ~/.zshrc && \
  pnpm add -g \
    "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
    "@openai/codex@${CODEX_VERSION}"

WORKDIR /home/${USERNAME}/projects
