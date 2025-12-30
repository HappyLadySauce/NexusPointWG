# WireGuard 管理平台实现完成度评估

本文档评估当前 WireGuard 管理平台的后端实现完成度，并规划需要补充的内容。

## 一、当前完成情况

### 1.1 Controller 层完成度

#### ✅ 已完成
- **CreatePeer** (`peer.go`)：创建 WireGuard Peer
  - ✅ 权限检查（Casbin）
  - ✅ 参数解析和校验
  - ✅ 调用 Service 层
  - ❌ **TODO**: 生成客户端配置文件
  - ❌ **TODO**: 更新服务器配置文件
  - ❌ **TODO**: 应用服务器配置

- **ListPeers** (`peer.go`)：列出 WireGuard Peers
  - ✅ 权限检查（Casbin）
  - ✅ 分页和过滤
  - ✅ 响应格式化

- **GetPeer** (`getPeer.go`)：获取单个 Peer
  - ✅ 权限检查（Casbin）
  - ✅ 响应格式化

- **DeletePeer** (`deletePeer.go`)：删除 Peer
  - ✅ 权限检查（Casbin）
  - ❌ **TODO**: Service 层释放 IP 分配
  - ❌ **TODO**: 更新服务器配置文件
  - ❌ **TODO**: 应用服务器配置

- **CreateIPPool** (`ip_pool.go`)：创建 IP 池（管理员）
  - ✅ 权限检查（Casbin）
  - ✅ 参数解析和校验

- **ListIPPools** (`ip_pool.go`)：列出 IP 池（管理员）
  - ✅ 权限检查（Casbin）
  - ✅ 分页和过滤

- **GetAvailableIPs** (`ip_pool.go`)：获取可用 IP 列表（管理员）
  - ✅ 权限检查（Casbin）
  - ❌ **TODO**: Service 层方法未实现

#### ❌ 缺失
- **UpdatePeer**：更新 Peer 配置
  - 需要实现 Controller、路由、权限检查

### 1.2 Service 层完成度

#### ✅ 已完成
- **WGPeerSrv.CreatePeer** (`wg_peer.go`)
  - ✅ IP 池自动选择（如果未指定）
  - ✅ IP 地址自动分配或验证指定 IP
  - ✅ 密钥对自动生成或验证指定密钥
  - ✅ 创建 Peer 记录
  - ✅ 创建 IP 分配记录
  - ✅ 事务处理（失败回滚）

- **WGPeerSrv.GetPeer**：获取 Peer
- **WGPeerSrv.GetPeerByPublicKey**：通过公钥获取 Peer
- **WGPeerSrv.UpdatePeer**：更新 Peer（简单实现，直接调用 Store）
- **WGPeerSrv.DeletePeer**：删除 Peer（简单实现，未释放 IP）
- **WGPeerSrv.ListPeers**：列出 Peers

- **IPPoolSrv**：IP 池的完整 CRUD 操作

#### ❌ 缺失/不完整
- **ReleaseIP**：释放 IP 分配
  - `ip.Allocator.ReleaseIP` 已实现，但 Service 层未封装
  - DeletePeer 需要调用此方法

- **GetAvailableIPs**：获取可用 IP 列表
  - `ip.Allocator.GetAvailableIPs` 已实现，但 Service 层未封装

- **UpdatePeer**：需要增强业务逻辑
  - 当前只是简单调用 Store，需要处理：
    - IP 地址变更时的重新分配
    - 状态变更时的配置同步
    - 参数校验和默认值设置

### 1.3 Store 层完成度

#### ✅ 已完成
- **WGPeerStore**：Peer 的完整 CRUD 操作
- **IPPoolStore**：IP 池的完整 CRUD 操作
- **IPAllocationStore**：IP 分配的完整 CRUD 操作
  - ✅ `GetAllocatedIPsByPoolID`：获取已分配 IP 列表

### 1.4 核心工具函数完成度

#### ✅ 已完成
- **密钥管理** (`internal/pkg/core/wireguard/keys.go`)
  - ✅ `GeneratePrivateKey`：生成私钥
  - ✅ `GeneratePublicKey`：从私钥生成公钥
  - ✅ `GenerateKeyPair`：生成密钥对
  - ✅ `ValidatePrivateKey`：验证私钥

- **配置生成** (`internal/pkg/core/wireguard/config.go`)
  - ✅ `GenerateClientConfig`：生成客户端配置文件内容
  - ✅ `FormatServerPeerBlock`：格式化服务器配置中的 Peer 块

- **IP 管理** (`internal/pkg/core/ip/allocator.go`)
  - ✅ `AllocateIP`：分配 IP 地址
  - ✅ `ValidateAndAllocateIP`：验证并分配指定 IP
  - ✅ `GetAvailableIPs`：获取可用 IP 列表
  - ✅ `ReleaseIP`：释放 IP 分配

#### ❌ 缺失
- **服务器配置文件管理**
  - 读取服务器配置文件
  - 解析服务器配置文件（Interface 和 Peer 块）
  - 更新服务器配置文件（添加/删除/更新 Peer）
  - 应用服务器配置（通过 systemctl 或 wg 命令）

- **客户端配置文件管理**
  - 生成并保存客户端配置文件到文件系统
  - 提供配置文件下载接口

### 1.5 权限模型完成度

#### ✅ 已完成
- **资源定义** (`internal/pkg/spec/spec.go`)
  - ✅ `ResourceWGPeer`
  - ✅ `ResourceIPPool`
  - ✅ `ResourceWGConfig`

- **操作定义**
  - ✅ `ActionWGPeerCreate`
  - ✅ `ActionWGPeerUpdate`
  - ✅ `ActionWGPeerDelete`
  - ✅ `ActionWGPeerList`
  - ✅ `ActionIPPoolCreate`
  - ✅ `ActionIPPoolUpdate`
  - ✅ `ActionIPPoolDelete`
  - ✅ `ActionIPPoolList`
  - ✅ `ActionWGConfigDownload`
  - ✅ `ActionWGConfigRotate`
  - ✅ `ActionWGConfigRevoke`
  - ✅ `ActionWGConfigUpdate`

### 1.6 路由注册

#### ✅ 已完成
- WireGuard 路由已注册 (`cmd/app/routes/wg/handler.go`)
- 但路由文件被注释（`cmd/app/api.go` 第 21 行）

## 二、需要补充的核心功能

### 2.1 配置文件管理服务（高优先级）

#### 2.1.1 服务器配置文件管理
需要实现一个服务来管理 WireGuard 服务器配置文件：

**功能需求**：
1. **读取服务器配置**
   - 读取 `/etc/wireguard/wg0.conf`（或配置的路径）
   - 解析 Interface 部分（PrivateKey, Address, ListenPort 等）
   - 解析 Peer 部分（PublicKey, AllowedIPs 等）

2. **更新服务器配置**
   - 添加新的 Peer 到配置文件
   - 删除 Peer（根据 PublicKey）
   - 更新 Peer 配置（AllowedIPs, PersistentKeepalive 等）
   - 保持配置文件格式和注释

3. **应用服务器配置**
   - 根据 `ApplyMethod` 配置：
     - `systemctl`：执行 `systemctl reload wg-quick@wg0`
     - `none`：仅更新文件，不应用
   - 错误处理和回滚机制

**实现建议**：
- 创建 `internal/pkg/core/wireguard/server_config.go`
- 实现 `ServerConfigManager` 结构体
- 提供方法：
  - `ReadServerConfig()`：读取并解析服务器配置
  - `AddPeer()`：添加 Peer
  - `RemovePeer()`：删除 Peer
  - `UpdatePeer()`：更新 Peer
  - `WriteServerConfig()`：写入配置文件
  - `ApplyConfig()`：应用配置

#### 2.1.2 客户端配置文件管理
需要实现客户端配置文件的生成和下载：

**功能需求**：
1. **生成客户端配置**
   - 使用 `GenerateClientConfig` 生成配置内容
   - 保存到配置的 `UserDir` 目录
   - 文件名建议：`{peerID}.conf` 或 `{username}-{deviceName}.conf`

2. **下载客户端配置**
   - 提供下载接口 `GET /api/v1/wg/peers/{id}/config`
   - 权限检查：用户只能下载自己的配置，管理员可以下载所有配置
   - 返回配置文件内容（Content-Type: text/plain）

**实现建议**：
- 在 Controller 层添加 `DownloadPeerConfig` 方法
- 在 Service 层添加 `GenerateAndSaveClientConfig` 方法
- 使用 `pkg/config/config.Get().WireGuard` 获取配置路径

### 2.2 UpdatePeer 接口（高优先级）

**功能需求**：
- 更新 Peer 的以下字段：
  - `DeviceName`
  - `AllowedIPs`
  - `DNS`
  - `Endpoint`
  - `PersistentKeepalive`
  - `Status`（active/disabled）

**实现要点**：
1. **Controller 层** (`updatePeer.go`)
   - 权限检查（Casbin `ActionWGPeerUpdate`）
   - 参数解析和校验
   - 调用 Service 层

2. **Service 层增强**
   - 业务逻辑校验：
     - Status 变更时，需要同步更新服务器配置
     - AllowedIPs 变更时，需要更新服务器配置
   - 调用配置文件管理服务更新服务器配置

3. **路由注册**
   - `PUT /api/v1/wg/peers/:id`

### 2.3 IP 管理增强（中优先级）

#### 2.3.1 Service 层封装
- 在 `WGPeerSrv` 中添加 `ReleaseIP` 方法
- 在 `IPPoolSrv` 中添加 `GetAvailableIPs` 方法
- 在 `DeletePeer` 中调用 `ReleaseIP`

#### 2.3.2 IP 池管理增强
- **UpdateIPPool**：更新 IP 池（修改 CIDR、ServerIP 等）
- **DeleteIPPool**：删除 IP 池（需要检查是否有已分配的 IP）
- **GetIPPool**：获取 IP 池详情（包含统计信息：总 IP 数、已分配数、可用数）

### 2.4 配置同步机制（中优先级）

**问题**：当前数据库中的 Peer 状态可能与实际服务器配置文件不同步。

**解决方案**：
1. **启动时同步**：应用启动时，读取服务器配置文件，与数据库同步
2. **定期同步**：定时任务检查配置一致性
3. **手动同步**：提供管理员接口手动触发同步

**实现建议**：
- 创建 `internal/service/wg_sync.go`
- 实现 `SyncServerConfigToDatabase` 方法
- 在应用启动时调用（可选）

### 2.5 统计和监控（低优先级）

**功能需求**：
1. **Peer 统计**
   - 总 Peer 数
   - 活跃 Peer 数
   - 按用户统计
   - 按 IP 池统计

2. **IP 池统计**
   - 总 IP 数
   - 已分配 IP 数
   - 可用 IP 数
   - 使用率

3. **网络流量统计**（可选）
   - 需要集成 WireGuard 的统计接口
   - 每个 Peer 的上传/下载流量

**实现建议**：
- 创建 `internal/controller/wireguard/stats.go`
- 提供 `GET /api/v1/wg/stats` 接口（管理员）

## 三、架构设计建议

### 3.1 配置文件管理服务设计

```go
// internal/pkg/core/wireguard/server_config.go

type ServerConfigManager struct {
    configPath string
    applyMethod string
}

// ServerConfig 表示服务器配置文件的完整结构
type ServerConfig struct {
    Interface *InterfaceConfig
    Peers     []*ServerPeerConfig
}

// InterfaceConfig 表示服务器 Interface 配置
type InterfaceConfig struct {
    PrivateKey  string
    Address     string
    ListenPort  int
    // ... 其他字段
}

// 方法：
// - ReadServerConfig() (*ServerConfig, error)
// - WriteServerConfig(config *ServerConfig) error
// - AddPeer(peer *ServerPeerConfig) error
// - RemovePeer(publicKey string) error
// - UpdatePeer(publicKey string, peer *ServerPeerConfig) error
// - ApplyConfig() error
```

### 3.2 Service 层集成

在 `WGPeerSrv` 中集成配置文件管理：

```go
type wgPeerSrv struct {
    store store.Factory
    configManager *wireguard.ServerConfigManager  // 新增
}

func (w *wgPeerSrv) CreatePeer(...) (*model.WGPeer, error) {
    // ... 现有逻辑
    
    // 生成客户端配置
    clientConfig := &wireguard.ClientConfig{...}
    configContent := wireguard.GenerateClientConfig(clientConfig)
    // 保存客户端配置
    
    // 更新服务器配置
    serverPeer := &wireguard.ServerPeerConfig{
        PublicKey: peer.ClientPublicKey,
        AllowedIPs: peer.ClientIP,
        // ...
    }
    if err := w.configManager.AddPeer(serverPeer); err != nil {
        // 回滚
    }
    
    // 应用服务器配置
    if err := w.configManager.ApplyConfig(); err != nil {
        // 记录错误，但不影响数据库操作
    }
    
    return peer, nil
}
```

### 3.3 错误处理策略

**配置文件操作失败时的处理**：
1. **创建 Peer 时**：
   - 数据库操作成功，但配置文件操作失败
   - 方案：记录错误日志，返回警告，但不阻止操作
   - 提供手动同步接口

2. **删除 Peer 时**：
   - 先删除数据库记录，再更新配置文件
   - 如果配置文件更新失败，记录错误，但不回滚数据库

3. **更新 Peer 时**：
   - 先更新数据库，再更新配置文件
   - 如果配置文件更新失败，记录错误

## 四、实现优先级

### 高优先级（核心功能）
1. ✅ **UpdatePeer Controller 和路由**
2. ✅ **服务器配置文件管理服务**（读取、更新、应用）
3. ✅ **客户端配置文件生成和下载接口**
4. ✅ **Service 层 IP 释放和获取可用 IP**

### 中优先级（增强功能）
5. ✅ **配置同步机制**（启动时同步）
6. ✅ **IP 池管理增强**（Update、Delete、统计）

### 低优先级（可选功能）
7. ⚪ **统计和监控接口**
8. ⚪ **网络流量统计**

## 五、注意事项

### 5.1 权限控制
- 所有配置文件操作都需要权限检查
- 下载配置文件需要 `ActionWGConfigDownload` 权限
- 更新服务器配置需要管理员权限或特殊权限

### 5.2 并发安全
- 配置文件读写需要加锁，避免并发修改
- 使用文件锁（`flock`）或内存锁（`sync.Mutex`）

### 5.3 配置一致性
- 数据库和配置文件需要保持一致性
- 提供检查和修复工具
- 记录配置变更日志

### 5.4 错误恢复
- 配置文件损坏时的恢复机制
- 备份配置文件
- 提供配置重置接口

## 六、关键设计考虑

### 6.1 服务器公钥管理

**问题**：生成客户端配置时需要服务器的公钥（`ClientConfig.PublicKey`），但当前配置选项中未存储。

**解决方案**：
1. **从服务器配置文件读取**
   - 读取服务器配置文件的 `[Interface]` 部分
   - 获取 `PrivateKey`
   - 使用 `wireguard.GeneratePublicKey(serverPrivateKey)` 生成公钥
   - 缓存服务器公钥（避免频繁读取文件）

2. **配置选项扩展**（可选）
   - 在 `WireGuardOptions` 中添加 `ServerPublicKey` 字段
   - 允许手动配置服务器公钥（用于只读配置文件场景）

**实现建议**：
- 在 `ServerConfigManager` 中实现 `GetServerPublicKey()` 方法
- 在生成客户端配置时调用此方法获取服务器公钥

### 6.2 默认值处理

**问题**：创建 Peer 时，某些字段可以使用服务器默认值（如 DNS、Endpoint、AllowedIPs）。

**当前实现**：
- `CreatePeer` 中，如果请求未提供这些字段，使用空值
- 需要在生成客户端配置时使用服务器默认值

**解决方案**：
- 在生成客户端配置时，如果 Peer 的字段为空，使用 `WireGuardOptions` 中的默认值：
  - `DNS` → `WireGuardOptions.DNS`
  - `Endpoint` → `WireGuardOptions.Endpoint`
  - `AllowedIPs` → `WireGuardOptions.DefaultAllowedIPs`

### 6.3 路由管理（未来扩展）

**当前设计**：每个 Peer 的 `AllowedIPs` 字段存储允许的路由。

**未来扩展考虑**：
1. **路由模板**：定义常用的路由模板（如 `0.0.0.0/0`、`::/0` 等）
2. **路由组**：将多个路由组织成组，便于管理
3. **路由策略**：根据用户角色或 IP 池自动分配路由

### 6.4 多服务器支持（未来扩展）

**当前设计**：假设只有一个 WireGuard 服务器。

**未来扩展考虑**：
1. **服务器管理**：支持多个 WireGuard 服务器
2. **负载均衡**：将 Peer 分配到不同的服务器
3. **故障转移**：服务器故障时的自动切换

### 6.5 配置备份和恢复

**建议**：
1. **自动备份**：每次修改服务器配置文件前自动备份
2. **版本管理**：保存配置文件的历史版本
3. **恢复机制**：提供配置文件恢复接口

## 七、总结

当前实现已经完成了基础的 CRUD 操作和权限控制，核心缺失的是：

1. **配置文件管理**：服务器配置文件的读取、更新、应用
2. **UpdatePeer 接口**：更新 Peer 配置
3. **客户端配置下载**：提供配置文件下载功能
4. **IP 管理增强**：Service 层封装和 DeletePeer 时的 IP 释放
5. **服务器公钥管理**：从配置文件读取服务器公钥用于生成客户端配置

建议按照优先级逐步实现，先完成高优先级功能，确保核心流程可用，再逐步增强功能。

### 实现顺序建议

1. **第一阶段**（核心功能）
   - 实现服务器配置文件管理服务（读取、解析、更新）
   - 实现 UpdatePeer Controller 和路由
   - 在 CreatePeer/DeletePeer/UpdatePeer 中集成配置文件管理

2. **第二阶段**（完善功能）
   - 实现客户端配置文件生成和下载
   - 实现 Service 层 IP 释放和获取可用 IP
   - 实现配置同步机制

3. **第三阶段**（增强功能）
   - IP 池管理增强
   - 统计和监控接口
   - 配置备份和恢复

