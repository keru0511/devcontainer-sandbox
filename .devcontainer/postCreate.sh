#!/bin/bash
set -euo pipefail

# Setup SSH keys if mounted
if [ -d "$HOME/.ssh-host" ]; then
  mkdir -p ~/.ssh
  cp "$HOME"/.ssh-host/* ~/.ssh/ 2>/dev/null || true
  chmod 700 ~/.ssh
  chmod 600 ~/.ssh/* 2>/dev/null || true
fi

# Setup gitconfig if mounted
[ -f "$HOME/.gitconfig-host" ] && cp "$HOME/.gitconfig-host" ~/.gitconfig || true
