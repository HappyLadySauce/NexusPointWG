# 1Panel 安装指南

本文档介绍如何在 1Panel 中安装和部署 NexusPointWG。

## 前置要求

- 已安装并运行 1Panel
- 已安装并运行 WireGuard 服务器
- WireGuard 配置文件位于 `/etc/wireguard/wg0.conf`（或自定义路径）

## 安装方式

### 方式一：通过安装脚本（推荐）

1. **下载并解压应用文件**

   在 1Panel 宿主机上执行以下命令：

   ```bash
   wget -O /tmp/nexuspointwg.tar.gz https://raw.githubusercontent.com/HappyLadySauce/NexusPointWG/refs/heads/main/docker/1panel/nexuspointwg.tar.gz
   
   tar -zxvf /tmp/nexuspointwg.tar.gz -C /opt/1panel/apps/local
   ```

   > **注意**：如果 GitHub 访问受限，可以手动下载 `docker/1panel/nexuspointwg.tar.gz` 文件并上传到服务器，然后执行解压命令。

2. **更新应用商店**

   - 登录 1Panel 管理界面
   - 进入"应用商店" → "本地应用"
   - 点击"更新应用商店"按钮
   - 等待更新完成

3. **安装应用**

   - 在应用商店中搜索 "NexusPointWG"
   - 点击"安装"按钮
   - 配置以下参数：
     - **监听端口**：Web 界面访问端口（默认：51830）
   - 点击"确认"开始安装

4. **配置 WireGuard 参数**

   安装完成后，需要在应用设置中配置 WireGuard 相关参数：

   - 进入应用详情页面
   - 点击"设置"或"环境变量"
   - 添加以下必需参数：
     - `WIREGUARD_ENDPOINT`：服务器公网 IP 和端口（如：`123.45.67.89:51820`）
     - `JWT_SECRET`：JWT 密钥，用于 Token 加密（建议使用随机字符串）

   > **提示**：可以通过 1Panel 的"终端"功能执行以下命令生成随机 JWT Secret：
   > ```bash
   > openssl rand -base64 32
   > ```

### 方式二：手动安装

1. **准备应用文件**

   从项目仓库下载或克隆应用文件：

   ```bash
   git clone https://github.com/HappyLadySauce/NexusPointWG.git
   cd NexusPointWG
   make 1panel  # 打包 1Panel 应用
   ```

   打包完成后，文件位于 `docker/1panel/nexuspointwg.tar.gz`

2. **上传到 1Panel**

   ```bash
   # 将文件上传到服务器后，解压到 1Panel 应用目录
   tar -zxvf nexuspointwg.tar.gz -C /opt/1panel/apps/local
   ```

3. **后续步骤**

   按照"方式一"的步骤 2-4 完成安装和配置。

## 配置说明

### 必需配置

| 参数 | 说明 | 示例 |
|------|------|------|
| `WIREGUARD_ENDPOINT` | 服务器公网 IP 和端口 | `123.45.67.89:51820` |
| `JWT_SECRET` | JWT 密钥，用于 Token 加密 | 随机字符串（建议 32 位以上） |

### 可选配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `WIREGUARD_ROOT_DIR` | WireGuard 配置目录 | `/etc/wireguard` |
| `WIREGUARD_INTERFACE` | 服务器接口名称 | `wg0` |
| `WIREGUARD_DNS` | 默认 DNS 服务器 | - |
| `WIREGUARD_DEFAULT_ALLOWED_IPS` | 默认 AllowedIPs | `0.0.0.0/0` |
| `WIREGUARD_APPLY_METHOD` | 配置应用方式 | `none` |
| `SQLITE_DATABASE` | 数据库文件路径 | `NexusPointWG.db` |
| `JWT_TIMEOUT` | Token 过期时间 | `24h` |

### 配置应用方式

`WIREGUARD_APPLY_METHOD` 参数说明：

- `systemctl`：自动执行 `systemctl reload wg-quick@wg0` 使配置立即生效
- `none`：仅更新配置文件，需要手动重载

## 目录权限设置

安装脚本会自动设置 WireGuard 目录权限：

- 安装时：将 `/etc/wireguard` 目录权限设置为 `51830:51830`（容器运行用户）
- 卸载时：恢复目录权限为 `root:root`

> **注意**：如果使用自定义 WireGuard 配置目录，需要手动设置权限：
> ```bash
> sudo chown -R 51830:51830 /your/wireguard/path
> sudo chmod 755 /your/wireguard/path
> ```

## 访问 Web 界面

安装完成后，在浏览器中访问：

- **Web 界面**：`http://your-server-ip:51830`（端口为安装时配置的监听端口）
- **API 文档**：`http://your-server-ip:51830/swagger/index.html`

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

### 方式一：通过 1Panel 升级

1. 在应用商店中查看是否有新版本
2. 点击"升级"按钮
3. 等待升级完成

### 方式二：手动升级

1. 下载新版本的应用文件
2. 解压到 `/opt/1panel/apps/local` 目录
3. 在 1Panel 中更新应用商店
4. 在应用详情页面点击"升级"

## 卸载应用

1. 在 1Panel 应用列表中找到 NexusPointWG
2. 点击"卸载"按钮
3. 确认卸载

> **注意**：卸载脚本会自动恢复 WireGuard 目录权限。

## 常见问题

### Q: 安装后无法访问 Web 界面？

A: 请检查：
1. 容器是否正常运行（在 1Panel 中查看容器状态）
2. 监听端口是否正确配置
3. 防火墙是否开放了对应端口
4. 查看容器日志排查错误

### Q: 如何查看容器日志？

A: 在 1Panel 应用详情页面，点击"日志"标签页即可查看容器日志。

### Q: 如何修改配置参数？

A: 在应用详情页面，点击"设置"或"环境变量"，可以添加或修改配置参数。修改后需要重启应用才能生效。

### Q: 配置文件更新后如何生效？

A: 有两种方式：
1. **自动生效**：设置 `WIREGUARD_APPLY_METHOD=systemctl`，系统会自动重载配置
2. **手动生效**：使用默认设置，然后在 1Panel 终端中执行：
   ```bash
   sudo systemctl reload wg-quick@wg0
   ```

### Q: 如何备份数据？

A: 系统使用 SQLite 数据库，数据库文件位于 WireGuard 配置目录下。可以通过 1Panel 的文件管理功能备份该文件。

## 相关资源

- [1Panel 应用创建教程](https://bbs.fit2cloud.com/t/topic/7409)
- [1Panel 第三方应用商店文档](https://doc.theojs.cn/notes/1panel-third-party-app-store)
- [项目 GitHub 仓库](https://github.com/HappyLadySauce/NexusPointWG)
- [Docker 安装文档](docker.md)
- [开发指南](../dev/dev.md)
