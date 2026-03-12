# 本地部署

这个项目现在的部署方式是：

- 在你的电脑上编译后端和前端
- 通过 `ssh + rsync` 上传到服务器
- 服务器只负责覆盖文件和重启 `systemd`

默认目标目录：

- 后端二进制：`/opt/agent-tracker`
- 前端静态文件：`/var/www/agent-tracker`
- 配置文件：`/etc/agent-tracker/config.toml`
- 数据目录：`/var/lib/agent-tracker`
- 日志文件：`/var/lib/agent-tracker/logs/sync.log`
- 服务名：`agent-tracker`
- 前端挂载路径：`/agent-tracker/`

## 前提

你的电脑需要能直接执行这些命令：

- `ssh`
- `rsync`
- `go`
- `bun`

服务器需要是 `linux/amd64`，也就是你说的 `x86_64`。

## 服务器准备

创建目录：

```bash
sudo mkdir -p /opt/agent-tracker /var/www/agent-tracker /etc/agent-tracker /var/lib/agent-tracker
```

创建配置文件 `/etc/agent-tracker/config.toml`：

```toml
data_dir = "/var/lib/agent-tracker"
log_path = "/var/lib/agent-tracker/logs/sync.log"
port = "10001"
```

创建 `systemd` 服务 `/etc/systemd/system/agent-tracker.service`：

```ini
[Unit]
Description=Agent Tracker
After=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/opt/agent-tracker
ExecStart=/opt/agent-tracker/agent-tracker --config /etc/agent-tracker/config.toml
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable agent-tracker
sudo systemctl start agent-tracker
```

## sudo 权限

部署用户需要能通过 `sudo` 执行这些命令：

- `/usr/bin/install`
- `/usr/bin/rsync`
- `/usr/bin/test`
- `/bin/systemctl daemon-reload`
- `/bin/systemctl restart agent-tracker`
- `/bin/systemctl status agent-tracker --no-pager`

示例 sudoers：

```bash
deploy ALL=(root) NOPASSWD: /usr/bin/install, /usr/bin/rsync, /usr/bin/test, /bin/systemctl daemon-reload, /bin/systemctl restart agent-tracker, /bin/systemctl status agent-tracker --no-pager
```

## 本地执行部署

部署脚本是 `scripts/deploy-local.sh`。

如果你本地 SSH 已经配好，可以直接这样跑：

```bash
./scripts/deploy-local.sh deploy@your-server-ip
```

如果你用的是 SSH config 里的别名：

```bash
./scripts/deploy-local.sh my-server
```

如果你想改服务器临时上传目录，可以传第二个参数：

```bash
./scripts/deploy-local.sh my-server /tmp/agent-tracker-custom-release
```

脚本会做这些事：

- 本地编译 Linux `amd64` 后端
- 本地构建挂载到 `/agent-tracker/` 的前端静态文件
- 上传到服务器临时目录
- 在服务器上覆盖正式目录
- 重启 `agent-tracker`

如果你跑完还不会用，那就说明不是脚本的问题，是你命令敲得像在打麻将。
