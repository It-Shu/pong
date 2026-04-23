#!/usr/bin/env sh
set -eu

REPO="It-Shu/pong"
BINARY_NAME="pong-terminal"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "Unsupported OS: $uname_s" >&2
    exit 1
    ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $uname_m" >&2
    exit 1
    ;;
esac

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

if command -v curl >/dev/null 2>&1; then
  release_json="$(curl -fsSL "$API_URL")"
elif command -v wget >/dev/null 2>&1; then
  release_json="$(wget -qO- "$API_URL")"
else
  echo "curl or wget is required" >&2
  exit 1
fi

version="$(printf '%s' "$release_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
if [ -z "$version" ]; then
  echo "Could not resolve latest release version" >&2
  exit 1
fi

asset="${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
download_url="https://github.com/${REPO}/releases/download/${version}/${asset}"
archive_path="$tmpdir/${asset}"

echo "Downloading ${asset}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$download_url" -o "$archive_path"
else
  wget -qO "$archive_path" "$download_url"
fi

mkdir -p "$INSTALL_DIR"
tar -xzf "$archive_path" -C "$tmpdir"
install -m 755 "$tmpdir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

path_line='export PATH="$HOME/.local/bin:$PATH"'
shell_name="${SHELL##*/}"
rc_file=""

case "$shell_name" in
  zsh) rc_file="$HOME/.zshrc" ;;
  bash) rc_file="$HOME/.bashrc" ;;
esac

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    if [ -n "$rc_file" ]; then
      if [ ! -f "$rc_file" ] || ! grep -F "$path_line" "$rc_file" >/dev/null 2>&1; then
        printf '\n%s\n' "$path_line" >> "$rc_file"
        echo "Added $INSTALL_DIR to PATH in $rc_file"
      fi
    fi
    export PATH="$INSTALL_DIR:$PATH"
    ;;
esac

echo
echo "Installed to: $INSTALL_DIR/$BINARY_NAME"
echo "Run: $BINARY_NAME"
