#!/bin/sh
set -e

REPO="eugenioenko/wp"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux)
    BINARY="wp-linux-${ARCH}"
    INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
    ;;
  darwin)
    BINARY="wp-darwin-${ARCH}"
    INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
    ;;
  *)
    echo "Unsupported OS: $OS"; exit 1
    ;;
esac

if [ -n "$1" ]; then
  VERSION="$1"
else
  VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

STAGING=$(mktemp -d)
trap 'rm -rf "$STAGING"' EXIT

echo "Downloading wp ${VERSION} for ${OS}/${ARCH}..."
curl -sSfL "$URL" -o "$STAGING/wp"

chmod +x "$STAGING/wp"

if [ "$OS" = "darwin" ]; then
  xattr -d com.apple.quarantine "$STAGING/wp" 2>/dev/null || true
fi

mkdir -p "$INSTALL_DIR"

if [ -w "$INSTALL_DIR" ]; then
  mv "$STAGING/wp" "${INSTALL_DIR}/wp"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$STAGING/wp" "${INSTALL_DIR}/wp"
fi

echo "wp ${VERSION} installed to ${INSTALL_DIR}/wp"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    SHELL_NAME=$(basename "${SHELL:-/bin/sh}")
    case "$SHELL_NAME" in
      zsh)  RC_FILE="~/.zshrc" ;;
      bash) RC_FILE="~/.bashrc" ;;
      fish) RC_FILE="~/.config/fish/config.fish" ;;
      *)    RC_FILE="~/.profile" ;;
    esac
    echo ""
    echo "To add wp to your PATH, add this to ${RC_FILE}:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
    echo "Then restart your terminal."
    ;;
esac
