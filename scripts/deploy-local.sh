#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <ssh_target> [release_dir]" >&2
  exit 1
fi

ssh_target="$1"
release_dir="${2:-/tmp/agent-tracker-release}"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "$script_dir/.." && pwd)"
build_dir="$repo_dir/build"

backend_bin_dir="/opt/agent-tracker"
frontend_dir="/var/www/agent-tracker"
config_dir="/etc/agent-tracker"
service_name="agent-tracker"
base_path="${AGENT_TRACKER_BASE_PATH:-/agent-tracker/}"

echo "==> building backend"
mkdir -p "$build_dir/backend"
pushd "$repo_dir/server" >/dev/null
GOOS=linux GOARCH=amd64 go build -o "$build_dir/backend/agent-tracker" .
popd >/dev/null

echo "==> building frontend"
pushd "$repo_dir/web" >/dev/null
bun install --frozen-lockfile
AGENT_TRACKER_BASE_PATH="$base_path" bun run build
popd >/dev/null

rm -rf "$build_dir/frontend"
mkdir -p "$build_dir/frontend"
cp -R "$repo_dir/web/dist/." "$build_dir/frontend/"

echo "==> uploading release to $ssh_target"
ssh "$ssh_target" "mkdir -p '$release_dir/backend' '$release_dir/frontend'"
rsync -avz --delete "$build_dir/backend/" "$ssh_target:$release_dir/backend/"
rsync -avz --delete "$build_dir/frontend/" "$ssh_target:$release_dir/frontend/"
rsync -avz "$repo_dir/scripts/deploy-remote.sh" "$ssh_target:$release_dir/"

echo "==> activating release on $ssh_target"
ssh "$ssh_target" \
  "chmod +x '$release_dir/deploy-remote.sh' && '$release_dir/deploy-remote.sh' '$release_dir' '$backend_bin_dir' '$frontend_dir' '$config_dir' '$service_name'"

echo "==> deploy complete"
