# 开发指南

本文档面向开发者，介绍如何搭建开发环境、构建项目以及进行开发调试。

## 开发环境要求

- Go 1.21 或更高版本
- Node.js 18 或更高版本
- pnpm 包管理器
- Docker 和 Docker Compose（可选，用于容器化开发）

## 获取项目

```bash
git clone https://github.com/HappyLadySauce/NexusPointWG.git
cd NexusPointWG
```

## 安装依赖

### 后端依赖

```bash
go mod tidy
```

### 前端依赖

```bash
cd ui
pnpm install
cd ..
```

## 本地开发运行

### 方式一：使用默认配置

```bash
make run
```

### 方式二：使用自定义配置

```bash
./bin/NexusPointWG \
  --wireguard.root-dir=/etc/wireguard \
  --wireguard.interface=wg0 \
  --wireguard.endpoint=your-server-ip:51820 \
  --jwt.secret=your-secret-key
```

**必需参数说明**：
- `--wireguard.endpoint`：您的服务器公网 IP 和端口（如：`123.45.67.89:51820`）
- `--jwt.secret`：JWT 密钥，用于 Token 加密（建议使用随机字符串）

### 访问开发服务

启动成功后，在浏览器中访问：

- **Web 界面**：http://localhost:51830
- **API 文档**：http://localhost:51830/swagger/index.html

## Docker 开发环境

### 构建开发镜像

项目已经集成版本管理，只需要通过 Make 命令即可构建带版本号的镜像：

```bash
# 构建开发环境镜像（tag 示例：1.0.1-dev）
make docker.build
# 等价于
make docker.build.dev
```

### 运行开发环境容器

开发环境使用 `docker-compose.dev.yml`，会将宿主机的 WireGuard 配置目录挂载到容器内 `/etc/wireguard`：

```bash
make docker.run          # 使用 dev 配置
# 等价于
make docker.run.dev
```

默认挂载目录（可在 `docker-compose.dev.yml` 中查看并修改）：

- 宿主机：`/opt/NexusPointWG/example/wireguard`
- 容器内：`/etc/wireguard`

**重要：请务必在宿主机上为该目录授予容器用户写权限，否则 WireGuard 配置和 SQLite 数据库无法创建/更新。**

容器内运行用户为 `NexusPointWG`，UID 为 `51830`，建议在宿主机执行：

```bash
sudo mkdir -p /opt/NexusPointWG/example/wireguard
sudo chown -R 51830:51830 /opt/NexusPointWG/example/wireguard
sudo chmod 755 /opt/NexusPointWG/example/wireguard
```

## 构建生产版本

### 构建二进制文件

```bash
make build
```

构建产物位于 `_output/` 目录。

### 构建生产 Docker 镜像

```bash
# 构建发布版本镜像（tag 示例：1.0.1）
make docker.build.release
```

## 版本管理

项目版本号统一在 `pkg/environment/version.go` 中管理：

- `dev`：开发版本号（如 `1.0.1-dev`）
- `release`：发布版本号（如 `1.0.1`）

构建时会自动从该文件读取版本号，并注入到二进制文件中。Docker 镜像标签也会使用对应的版本号。

## 推送镜像到 Docker Hub

项目内置 `make docker.push`，仅推送 release 版本镜像到 Docker Hub 用户 `happlelaoganma`：

```bash
# 先构建 release 镜像
make docker.build.release

# 登录 Docker Hub（只需执行一次）
docker login

# 推送镜像
make docker.push
```

该命令会自动从 `pkg/environment/version.go` 中读取 release 版本号，并推送：

- `happlelaoganma/nexuspointwg:<release-version>`（例如 `1.0.1`）
- `happlelaoganma/nexuspointwg:latest`

## 开发工具

### 代码格式化

```bash
# 格式化 Go 代码
go fmt ./...

# 格式化前端代码（如果配置了）
cd ui && pnpm format
```

### 运行测试

```bash
# 运行 Go 测试
go test ./...

# 运行前端测试（如果配置了）
cd ui && pnpm test
```

### 生成 API 文档

```bash
make swagger
```

## 项目结构

```
NexusPointWG/
├── cmd/                    # 应用程序入口
│   └── app/               # 主应用入口
├── internal/              # 内部包（不对外暴露）
│   ├── pkg/              # 内部核心包
│   │   ├── core/         # 核心业务逻辑
│   │   └── model/        # 数据模型
│   └── store/            # 数据存储层
├── pkg/                   # 公共包（可对外暴露）
│   ├── environment/      # 环境配置（版本号等）
│   ├── options/           # 配置选项
│   └── utils/            # 工具函数
├── ui/                    # 前端应用
├── configs/               # 配置文件
├── docs/                  # 文档
├── scripts/               # 构建脚本
└── Makefile               # Make 构建文件
```

详细的项目分层规范请参考 [项目分层规范](项目分层规范.md)。

## 常见开发问题

### Q: 如何修改 WireGuard 配置路径？

A: 使用 `--wireguard.root-dir` 参数指定配置目录，例如：
```bash
./bin/NexusPointWG --wireguard.root-dir=/custom/path/wireguard
```

### Q: 配置文件更新后如何生效？

A: 有两种方式：
1. **自动生效**：设置 `--wireguard.apply-method=systemctl`，系统会自动重载配置
2. **手动生效**：使用默认设置，然后手动执行 `sudo systemctl reload wg-quick@wg0`

### Q: 如何查看系统日志？

A: 日志默认输出到标准输出，可以通过重定向保存到文件：
```bash
./bin/NexusPointWG > nexuspointwg.log 2>&1
```

### Q: Docker 容器无法创建配置文件？

A: 检查挂载目录的权限，确保容器用户（UID 51830）有写权限。参考上方"运行开发环境容器"章节的权限设置说明。

## 相关文档

- [项目分层规范](项目分层规范.md) - 开发者架构设计文档
- [API 接口规范](api-contract.md) - API 接口对接规范
- [UI 设计规范](design/UI_SPECS.md) - 前端 UI 设计指南

