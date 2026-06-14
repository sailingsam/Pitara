#!/bin/sh
# Pitara installer — detects your OS/arch, downloads the right binary, installs it.
# Usage:  curl -fsSL https://pitara.dev/install.sh | sh
set -e

REPO="sailingsam/pitara"

# --- detect OS ---
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux)  os="linux" ;;
  darwin) os="darwin" ;;
  *)
    echo "Pitara: unsupported OS '$os'."
    echo "On Windows, install with: npm i -g pitara"
    echo "Or download a binary: https://github.com/$REPO/releases/latest"
    exit 1 ;;
esac

# --- detect architecture ---
arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Pitara: unsupported architecture '$arch'."
    echo "Download a binary: https://github.com/$REPO/releases/latest"
    exit 1 ;;
esac

asset="pitara_${os}_${arch}"
url="https://github.com/$REPO/releases/latest/download/$asset"

# --- pick an install dir (prefer /usr/local/bin, fall back to ~/.local/bin) ---
dir="/usr/local/bin"
if [ ! -d "$dir" ] || [ ! -w "$dir" ]; then
  dir="$HOME/.local/bin"
  mkdir -p "$dir"
fi

echo "Pitara: downloading $asset ..."
curl -fsSL "$url" -o "$dir/pitara"
chmod +x "$dir/pitara"
echo "Pitara: installed to $dir/pitara"

# --- PATH hint ---
case ":$PATH:" in
  *":$dir:"*) ;;
  *)
    echo ""
    echo "  $dir is not on your PATH. Add it:"
    echo "    export PATH=\"$dir:\$PATH\"     # add this to ~/.bashrc or ~/.zshrc"
    echo "" ;;
esac

echo "Done! Run:  pitara --help"
