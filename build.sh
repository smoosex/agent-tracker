#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<EOF
usage: $0 [backend|frontend|all]

  backend   Build server (linux/amd64) to build/agent-tracker
  frontend  Build web frontend and zip to build/frontend.zip
  all       Build both (default)

EOF
  exit 1
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
build_dir="$script_dir/build"
base_path="${AGENT_TRACKER_BASE_PATH:-/agent-tracker/}"

build_backend() {
  echo "==> building backend (linux/amd64)"
  mkdir -p "$build_dir"
  pushd "$script_dir/server" >/dev/null
  GOOS=linux GOARCH=amd64 go build -o "$build_dir/agent-tracker" .
  popd >/dev/null
  echo "    -> $build_dir/agent-tracker"
}

build_frontend() {
  echo "==> building frontend"
  pushd "$script_dir/web" >/dev/null
  bun install --frozen-lockfile
  AGENT_TRACKER_BASE_PATH="$base_path" bun run build
  popd >/dev/null

  echo "==> packaging frontend.zip"
  rm -f "$build_dir/frontend.zip"
  pushd "$script_dir/web/dist" >/dev/null
  zip -qr "$build_dir/frontend.zip" .
  popd >/dev/null
  echo "    -> $build_dir/frontend.zip"
}

case "${1:-all}" in
  backend)   build_backend ;;
  frontend)  build_frontend ;;
  all)       build_backend && build_frontend ;;
  *)         usage ;;
esac
