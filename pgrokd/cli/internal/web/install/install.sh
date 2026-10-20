#!/bin/sh
# Installs the pgrok client and runs it as a background service.
#
#   curl -fsSL __SERVER_URL__/install.sh | sh -s -- --token <token> [--forward-addr 3000] [--basic-auth user:password]
#
# macOS: per-user LaunchAgent (asks for sudo only to put the binary in /usr/local/bin).
# Linux: systemd unit running as you (needs sudo).
set -eu

REMOTE_ADDR="__REMOTE_ADDR__"
DOWNLOAD_URL="__DOWNLOAD_URL__"
FORWARD_ADDR="3000"
TOKEN=""
BASIC_AUTH=""
INSTALL_DIR="/usr/local/bin"

while [ $# -gt 0 ]; do
  case "$1" in
    --token|-t) TOKEN="$2"; shift 2 ;;
    --forward-addr|-f) FORWARD_ADDR="$2"; shift 2 ;;
    --basic-auth|-a) BASIC_AUTH="$2"; shift 2 ;;
    --remote-addr|-r) REMOTE_ADDR="$2"; shift 2 ;;
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done
[ -n "$TOKEN" ] || { echo "--token is required (copy it from __SERVER_URL__)" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "unsupported OS: $(uname -s) — use install.ps1 on Windows" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

sudo=""
if [ "$(id -u)" -ne 0 ]; then
  command -v sudo >/dev/null || { echo "sudo is required" >&2; exit 1; }
  sudo="sudo"
fi

echo "==> Downloading pgrok ($os/$arch)"
tmp="$(mktemp)"
curl -fsSL -o "$tmp" "$DOWNLOAD_URL/pgrok_${os}_${arch}"
chmod +x "$tmp"
dirsudo="$sudo"
mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ] && dirsudo=""
$dirsudo mkdir -p "$INSTALL_DIR"
$dirsudo mv "$tmp" "$INSTALL_DIR/pgrok"
"$INSTALL_DIR/pgrok" --version

config="$HOME/.pgrok/pgrok.yml"
echo "==> Writing $config"
set -- init --config "$config" --remote-addr "$REMOTE_ADDR" --forward-addr "$FORWARD_ADDR" --token "$TOKEN"
[ -z "$BASIC_AUTH" ] || set -- "$@" --basic-auth "$BASIC_AUTH"
"$INSTALL_DIR/pgrok" "$@"

echo "==> Installing service"
# macOS: per-user LaunchAgent. Linux: system unit (runs as you via SUDO_USER).
svcsudo="$sudo"
[ "$os" = darwin ] && svcsudo=""
$svcsudo "$INSTALL_DIR/pgrok" service install --config "$config"

echo
echo "pgrok is running in the background and starts automatically."
echo "  status:    $INSTALL_DIR/pgrok service status"
echo "  stop:      ${svcsudo:+sudo }$INSTALL_DIR/pgrok service stop"
echo "  uninstall: ${svcsudo:+sudo }$INSTALL_DIR/pgrok service uninstall"
