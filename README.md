# Agent Tracker

[English](README.md) | [简体中文](README.zh-CN.md)

Agent Tracker is a small web app for tracking release notes and changelogs from AI coding tools.

It collects updates from GitHub Releases and official changelog pages, stores them in a local SQLite database, and exposes a simple UI for browsing recent releases, filtering by tool, searching historical entries, checking sync logs, and subscribing via RSS.

![Agent Tracker demo](assets/demo.png)

## Tracked Tools

- Claude Code
- Codex App
- Codex CLI
- Gemini CLI
- OpenCode
- Pi

## Project Structure

- `server`: Go + Gin + SQLite backend for syncing, storage, and APIs
- `web`: React + Vite + Tailwind frontend

## Local Development

### 1. Start the backend

```bash
cd server
mkdir -p data/logs
go run .
```

The backend reads `server/config.toml` by default and runs on port `10001`.

### 2. Start the frontend

Open another terminal:

```bash
cd web
bun install
bun run dev
```

The frontend runs on `http://localhost:20001` and proxies `/api` and `/rss` requests to `http://localhost:10001`.

### 3. Open the app

Visit:

```text
http://localhost:20001
```

## Configuration

Default backend config lives in `server/config.toml`:

```toml
data_dir = "./data"
log_path = "./data/logs/sync.log"
port = "10001"
```

Edit this file if you want to change the data directory, log path, or port.
