#!/usr/bin/env bash
# Compila uma versão nova do naHora e (opcional) publica no GitHub Releases.
# Uso: build/release.sh 1.2.0 "O que mudou"   [--publicar]
set -euo pipefail
VER="$1"; NOTES="${2:-}"; cd "$(dirname "$0")/.."
REPO="${NAHORA_REPO:-}"
python3 build/mkrsrc.py build rsrc_windows_amd64.syso "$VER"
GOOS=windows GOARCH=amd64 go build -trimpath \
  -ldflags "-H windowsgui -s -w -X main.appVersion=$VER -X main.defaultUpdateRepo=$REPO" \
  -o build/naHora.exe .
(cd build && sha256sum naHora.exe > naHora.exe.sha256)
echo "ok: build/naHora.exe v$VER"
if [[ "${3:-}" == "--publicar" ]]; then
  gh release create "v$VER" build/naHora.exe build/naHora.exe.sha256 -R "$REPO" -t "$NOTES" -n "$NOTES"
fi
