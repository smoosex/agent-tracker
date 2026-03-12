#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 5 ]]; then
  echo "usage: $0 <release_dir> <backend_bin_dir> <frontend_dir> <config_dir> <service_name>" >&2
  exit 1
fi

release_dir="$1"
backend_bin_dir="$2"
frontend_dir="$3"
config_dir="$4"
service_name="$5"

backend_src="$release_dir/backend/agent-tracker"
frontend_src="$release_dir/frontend"

if [[ ! -f "$backend_src" ]]; then
  echo "backend binary not found: $backend_src" >&2
  exit 1
fi

if ! sudo test -f "$config_dir/config.toml"; then
  echo "config not found: $config_dir/config.toml" >&2
  exit 1
fi

sudo install -d -m 755 "$backend_bin_dir"
sudo install -d -m 755 "$frontend_dir"

sudo install -m 755 "$backend_src" "$backend_bin_dir/agent-tracker"
sudo rsync -a --delete "$frontend_src/" "$frontend_dir/"

sudo systemctl daemon-reload
sudo systemctl restart "$service_name"
sudo systemctl status "$service_name" --no-pager
