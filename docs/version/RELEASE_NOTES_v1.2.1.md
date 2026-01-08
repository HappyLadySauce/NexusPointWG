# NexusPointWG v1.2.1 版本发布

## 🎉 版本概述

本次更新主要包含项目结构优化和新增 1Panel 应用商店支持，提升了项目的可维护性和部署便利性。

---

## ✨ 新增功能

### 1Panel 应用商店支持

NexusPointWG 现已支持通过 1Panel 应用商店一键安装和部署，大大简化了安装流程。

**功能特点**：
- 支持在 1Panel 应用商店中直接安装
- 自动配置容器和网络设置
- 支持自定义监听端口
- 完整的安装和卸载脚本
- 自动设置 WireGuard 目录权限

**使用场景**：
- 使用 1Panel 管理服务器的用户
- 需要快速部署 NexusPointWG 的场景
- 希望通过图形界面管理应用的用户

**安装方式**：
1. 在 1Panel 中进入"应用商店"
2. 选择"本地应用"或通过安装脚本添加
3. 搜索 "NexusPointWG" 并安装
4. 配置监听端口（默认 51830）
5. 一键启动

**相关资源**：
- [1Panel 应用创建教程](https://bbs.fit2cloud.com/t/topic/7409)
- [1Panel 第三方应用商店文档](https://doc.theojs.cn/notes/1panel-third-party-app-store)
- [详细安装文档](../install/1panel.md)

---

## 🔧 技术改进

### 项目结构优化

**Docker 文件重组**：
- 将根目录下的 Docker 相关文件统一移动到 `docker/` 目录
- 提升了项目结构的清晰度和可维护性
- 便于 Docker 相关资源的集中管理

**文件变更**：
- `Dockerfile` → `docker/Dockerfile`
- `docker-compose.dev.yml` → `docker/docker-compose.dev.yml`
- `docker-compose.release.yml` → `docker/docker-compose.release.yml`
- `.dockerignore` → `docker/.dockerignore`

**Makefile 优化**：
- 创建了独立的 `scripts/make-rules/docker.mk` 模块
- 简化了主 Makefile 的结构
- 所有 Docker 相关命令路径已自动更新
- 新增 `make 1panel` 命令用于打包 1Panel 应用

**构建流程改进**：
- 构建前自动清理 `_output` 目录
- 确保每次构建都是干净的环境
- 提升了构建的可靠性

---

## 📝 详细变更

### 文件结构变更

```
项目根目录/
├── docker/                          # 新增：Docker 相关文件目录
│   ├── Dockerfile                   # 从根目录移动
│   ├── docker-compose.dev.yml       # 从根目录移动
│   ├── docker-compose.release.yml   # 从根目录移动
│   ├── .dockerignore                # 从根目录移动
│   └── 1panel/                      # 新增：1Panel 应用文件
│       ├── install.sh
│       ├── nexuspointwg.tar.gz
│       └── nexuspointwg/
│           ├── data.yml
│           ├── logo.png
│           ├── README.md
│           └── 1.2.1/
│               ├── data.yml
│               ├── docker-compose.yml
│               └── scripts/
│                   ├── init.sh
│                   └── uninstall.sh
└── scripts/make-rules/
    └── docker.mk                    # 新增：Docker 构建规则模块
```

### 代码变更

**Makefile**：
- 添加 `include scripts/make-rules/docker.mk`
- 移除所有 Docker 相关目标（已迁移到 `docker.mk`）
- 新增 `make 1panel` 命令用于打包 1Panel 应用
- 构建前自动清理 `_output` 目录

**scripts/make-rules/docker.mk**（新建）：
- 包含所有 Docker 构建、运行、管理相关目标
- 使用 `$(ROOT_DIR)` 构建路径，保持模块化
- 支持开发版本和发布版本的构建
- 支持 Docker Compose 管理

**docker-compose 文件**：
- 更新 `build.context` 为 `..`（项目根目录）
- 更新 `build.dockerfile` 为 `docker/Dockerfile`
- 保持所有功能不变，仅路径调整

---

## 🔄 升级指南

### 从 v1.2.0 升级到 v1.2.1

#### Docker 部署用户

**使用 Docker Compose**：
1. 更新代码：
   ```bash
   git pull origin main
   ```

2. 重新构建镜像（如果使用本地构建）：
   ```bash
   make docker.build.release
   ```

3. 或直接使用 Docker Hub 镜像：
   ```bash
   docker pull happlelaoganma/nexuspointwg:1.2.1
   ```

4. 更新 docker-compose 文件路径：
   - 如果使用 `docker-compose.release.yml`，文件已移动到 `docker/` 目录
   - 更新命令为：`docker compose -f docker/docker-compose.release.yml up -d`

**使用 Docker 命令**：
- 无需变更，直接拉取新镜像即可：
  ```bash
  docker pull happlelaoganma/nexuspointwg:1.2.1
  docker stop nexuspointwg
  docker rm nexuspointwg
  # 使用新镜像重新运行（参数保持不变）
  ```

#### 1Panel 用户

1. 在 1Panel 中卸载旧版本（如果已安装）
2. 通过应用商店安装新版本
3. 或使用安装脚本：
   ```bash
   wget -O /tmp/nexuspointwg.tar.gz https://raw.githubusercontent.com/HappyLadySauce/NexusPointWG/refs/heads/main/docker/1panel/nexuspointwg.tar.gz
   tar -zxvf /tmp/nexuspointwg.tar.gz -C /opt/1panel/apps/local
   ```
4. 在 1Panel 中更新应用商店并安装

#### 开发者

1. 更新代码：
   ```bash
   git pull origin main
   ```

2. 注意文件路径变更：
   - Docker 相关文件已移动到 `docker/` 目录
   - Makefile 中的 Docker 命令路径已自动更新
   - 如需手动构建，使用 `make docker.build` 即可

3. 重新构建：
   ```bash
   make docker.build
   ```

---

## 📊 统计数据

- **新增功能**：1 项（1Panel 应用商店支持）
- **技术改进**：2 项（项目结构优化、Makefile 模块化）
- **文件变更**：10+ 个文件移动/新增
- **向后兼容**：✅ 完全兼容，无需数据迁移

---

## ⚠️ 注意事项

1. **Docker Compose 文件路径变更**：
   - 如果使用自定义脚本调用 docker-compose，需要更新文件路径
   - 从 `docker-compose.release.yml` 更新为 `docker/docker-compose.release.yml`

2. **构建脚本更新**：
   - 如果使用 CI/CD 脚本，确保更新 Docker 相关路径
   - Makefile 命令保持不变，内部路径已自动处理

3. **1Panel 安装**：
   - 首次安装需要设置 WireGuard 目录权限
   - 安装脚本会自动处理权限设置
   - 卸载时会恢复权限

---

## 🙏 致谢

感谢所有贡献者和用户的支持与反馈！

特别感谢 1Panel 社区提供的应用商店平台和文档支持。

---

## 📚 相关链接

- [完整变更日志](../../CHANGELOG.md)
- [版本管理文档](../VERSION_MANAGEMENT.md)
- [项目文档](../../README.md)
- [1Panel 安装文档](../install/1panel.md)
- [Docker 安装文档](../install/docker.md)
- [开发指南](../dev/dev.md)

---

**Full Changelog**: https://github.com/HappyLadySauce/NexusPointWG/compare/v1.2.0...v1.2.1
