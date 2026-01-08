# Docker 安装指南

本文档介绍如何使用 Docker 部署 NexusPointWG。推荐使用 Docker Compose 方式部署，更便于管理和维护。

## 前置要求

- 已安装 Docker 和 Docker Compose
- 已安装并运行 WireGuard 服务器
- WireGuard 配置文件位于 `/etc/wireguard/wg0.conf`（或自定义路径）

## 方式一：使用 Docker Compose（推荐）

Docker Compose 方式提供了更好的配置管理和服务编排能力，推荐在生产环境使用。

### 1. 拉取镜像

从 Docker Hub 拉取最新版本：

```bash
docker pull happlelaoganma/nexuspointwg:latest
```

或者拉取指定版本：

```bash
docker pull happlelaoganma/nexuspointwg:1.2.1
```

### 2. 准备配置文件

从项目仓库获取 docker-compose 配置文件：

```bash
# 克隆项目（如果还没有）
git clone https://github.com/HappyLadySauce/NexusPointWG.git
cd NexusPointWG
```

或者直接下载配置文件：

```bash
# 下载发布版本的 compose 文件
wget https://raw.githubusercontent.com/HappyLadySauce/NexusPointWG/main/docker/docker-compose.release.yml
```

### 3. 设置目录权限

容器内运行用户为 `NexusPointWG`，UID 为 `51830`。需要确保宿主机 `/etc/wireguard` 目录对容器用户可写（用于生成/更新配置与数据库）：

```bash
sudo chown -R 51830:51830 /etc/wireguard
sudo chmod 755 /etc/wireguard
```

> **注意**：请根据实际运维需要自行权衡是否直接修改 `/etc/wireguard` 的属主/权限，或改用单独的数据目录。

### 4. 配置环境变量

编辑 `docker-compose.release.yml` 文件，或通过环境变量设置镜像版本：

```bash
# 设置环境变量（例如：1.2.1 或 latest）
export IMAGE_TAG=1.2.1
```

如果需要修改其他配置（如端口、挂载目录等），可以直接编辑 `docker-compose.release.yml` 文件。

### 5. 运行容器

使用 Docker Compose 启动服务：

```bash
# 使用环境变量指定镜像版本
IMAGE_TAG=1.2.1 docker compose -f docker/docker-compose.release.yml up -d

# 或者使用 latest 版本
IMAGE_TAG=latest docker compose -f docker/docker-compose.release.yml up -d
```

### 6. 配置应用参数

容器启动后，需要配置 WireGuard 相关参数。可以通过以下方式：

**方式一：通过环境变量（推荐）**

编辑 `docker-compose.release.yml`，在 `services.nexuspointwg` 下添加 `environment` 或 `command` 部分：

```yaml
services:
  nexuspointwg:
    image: happlelaoganma/nexuspointwg:${IMAGE_TAG:-latest}
    # ... 其他配置 ...
    command:
      - -c
      - /app/configs/NexusPointWG.yaml
      - --wireguard.endpoint=your-server-ip:51820
      - --jwt.secret=your-secret-key
```

然后重新启动：

```bash
docker compose -f docker/docker-compose.release.yml up -d
```

**方式二：通过配置文件**

1. 创建配置文件 `configs/NexusPointWG.yaml`（如果不存在）
2. 在配置文件中设置参数
3. 将配置文件挂载到容器中

### 7. 查看日志

```bash
# 查看容器日志
docker compose -f docker/docker-compose.release.yml logs -f

# 或使用容器名称
docker logs -f nexuspointwg
```

### 8. 停止和重启

```bash
# 停止服务
docker compose -f docker/docker-compose.release.yml down

# 重启服务
docker compose -f docker/docker-compose.release.yml restart

# 停止并删除容器（保留数据）
docker compose -f docker/docker-compose.release.yml down

# 停止并删除容器和卷（删除所有数据）
docker compose -f docker/docker-compose.release.yml down -v
```

## 方式二：直接使用 Docker 命令

如果不想使用 Docker Compose，也可以直接使用 Docker 命令运行容器。

### 1. 拉取镜像

```bash
docker pull happlelaoganma/nexuspointwg:latest
```

### 2. 设置目录权限

```bash
sudo chown -R 51830:51830 /etc/wireguard
sudo chmod 755 /etc/wireguard
```

### 3. 运行容器

```bash
docker run -d \
  --name nexuspointwg \
  --privileged \
  --pid host \
  -p 51830:51830 \
  -v /etc/wireguard:/etc/wireguard:rw \
  --restart unless-stopped \
  happlelaoganma/nexuspointwg:latest \
  -c /app/configs/NexusPointWG.yaml \
  --wireguard.endpoint=your-server-ip:51820 \
  --jwt.secret=your-secret-key
```

**参数说明**：
- `--name nexuspointwg`：容器名称
- `--privileged`：授予容器特权模式（用于访问系统服务）
- `--pid host`：使用主机 PID 命名空间（用于访问 systemctl）
- `-p 51830:51830`：端口映射（宿主机端口:容器端口）
- `-v /etc/wireguard:/etc/wireguard:rw`：挂载 WireGuard 配置目录
- `--restart unless-stopped`：自动重启策略
- `--wireguard.endpoint`：服务器公网 IP 和端口（必需）
- `--jwt.secret`：JWT 密钥，用于 Token 加密（必需）

### 4. 查看日志

```bash
docker logs -f nexuspointwg
```

### 5. 停止和删除

```bash
# 停止容器
docker stop nexuspointwg

# 删除容器
docker rm nexuspointwg

# 停止并删除容器
docker rm -f nexuspointwg
```

## 必需配置参数

无论使用哪种方式，以下参数都是必需的：

| 参数 | 说明 | 示例 |
|------|------|------|
| `--wireguard.endpoint` | 服务器公网 IP 和端口 | `123.45.67.89:51820` |
| `--jwt.secret` | JWT 密钥，用于 Token 加密 | 随机字符串（建议 32 位以上） |

**生成随机 JWT Secret**：

```bash
openssl rand -base64 32
```

## 可选配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--wireguard.root-dir` | WireGuard 配置目录 | `/etc/wireguard` |
| `--wireguard.interface` | 服务器接口名称 | `wg0` |
| `--wireguard.dns` | 默认 DNS 服务器 | - |
| `--wireguard.default-allowed-ips` | 默认 AllowedIPs | `0.0.0.0/0` |
| `--wireguard.apply-method` | 配置应用方式 | `none` |
| `--sqlite.database` | 数据库文件路径 | `NexusPointWG.db` |
| `--jwt.timeout` | Token 过期时间 | `24h` |

### 配置应用方式

`--wireguard.apply-method` 参数说明：

- `systemctl`：自动执行 `systemctl reload wg-quick@wg0` 使配置立即生效
- `none`：仅更新配置文件，需要手动重载

## 访问 Web 界面

启动成功后，在浏览器中访问：

- **Web 界面**：http://your-server-ip:51830
- **API 文档**：http://your-server-ip:51830/swagger/index.html

## 首次使用

### 1. 注册管理员账号

首次使用时，访问注册页面创建管理员账号。第一个注册的用户将自动获得管理员权限。

### 2. 创建 IP 地址池

登录后，进入 IP 池管理页面，创建您的第一个 IP 地址池：

- **IP 池地址**：输入 CIDR 格式的 IP 段，如 `10.0.0.0/24`
- **描述**：可选，用于说明这个 IP 池的用途

### 3. 创建第一个设备（Peer）

进入设备管理页面，点击"创建设备"，填写设备信息并下载配置文件。

## 升级应用

### 使用 Docker Compose

```bash
# 拉取新版本镜像
docker pull happlelaoganma/nexuspointwg:1.2.1

# 更新环境变量
export IMAGE_TAG=1.2.1

# 重新启动
IMAGE_TAG=1.2.1 docker compose -f docker/docker-compose.release.yml up -d
```

### 使用 Docker 命令

```bash
# 停止并删除旧容器
docker rm -f nexuspointwg

# 拉取新版本镜像
docker pull happlelaoganma/nexuspointwg:1.2.1

# 使用新镜像重新运行（参数保持不变）
docker run -d \
  --name nexuspointwg \
  --privileged \
  --pid host \
  -p 51830:51830 \
  -v /etc/wireguard:/etc/wireguard:rw \
  --restart unless-stopped \
  happlelaoganma/nexuspointwg:1.2.1 \
  -c /app/configs/NexusPointWG.yaml \
  --wireguard.endpoint=your-server-ip:51820 \
  --jwt.secret=your-secret-key
```

## 数据备份

系统使用 SQLite 数据库，数据库文件位于 WireGuard 配置目录下。备份时只需备份该文件：

```bash
# 备份数据库
cp /etc/wireguard/NexusPointWG.db /backup/NexusPointWG.db.$(date +%Y%m%d)

# 备份整个 WireGuard 配置目录
tar -czf /backup/wireguard-$(date +%Y%m%d).tar.gz /etc/wireguard
```

## 常见问题

### Q: 容器启动失败，提示权限错误？

A: 确保 WireGuard 目录权限正确设置：
```bash
sudo chown -R 51830:51830 /etc/wireguard
sudo chmod 755 /etc/wireguard
```

### Q: 配置文件更新后如何生效？

A: 有两种方式：
1. **自动生效**：设置 `--wireguard.apply-method=systemctl`，系统会自动重载配置
2. **手动生效**：使用默认设置，然后执行：
   ```bash
   sudo systemctl reload wg-quick@wg0
   ```

### Q: 如何修改 WireGuard 配置路径？

A: 使用 `--wireguard.root-dir` 参数指定配置目录，并在挂载卷时使用相同路径：
```bash
-v /custom/path/wireguard:/custom/path/wireguard:rw
```

### Q: 如何查看容器日志？

A: 
```bash
# Docker Compose
docker compose -f docker/docker-compose.release.yml logs -f

# Docker 命令
docker logs -f nexuspointwg
```

### Q: 容器无法访问 systemctl？

A: 确保使用了 `--privileged` 和 `--pid host` 参数。如果仍然无法访问，可能需要检查主机的 systemd 配置。

## 相关资源

- [1Panel 安装文档](1panel.md)
- [开发指南](../dev/dev.md)
- [项目 GitHub 仓库](https://github.com/HappyLadySauce/NexusPointWG)
- [Docker Hub 镜像](https://hub.docker.com/r/happlelaoganma/nexuspointwg)
