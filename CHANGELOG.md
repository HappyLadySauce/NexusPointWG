# Changelog

所有重要的项目变更都会记录在这个文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [1.2.1] - 2025-01-XX

### 新增功能
- **1Panel 应用商店支持**：支持通过 1Panel 应用商店一键安装和部署 NexusPointWG
  - 完整的安装和卸载脚本
  - 自动配置容器和网络设置
  - 支持自定义监听端口
  - 自动设置 WireGuard 目录权限

### 技术改进
- **项目结构优化**：将根目录下的 Docker 相关文件统一移动到 `docker/` 目录
  - `Dockerfile` → `docker/Dockerfile`
  - `docker-compose.dev.yml` → `docker/docker-compose.dev.yml`
  - `docker-compose.release.yml` → `docker/docker-compose.release.yml`
  - `.dockerignore` → `docker/.dockerignore`
- **Makefile 模块化**：创建独立的 `scripts/make-rules/docker.mk` 模块
  - 简化了主 Makefile 的结构
  - 所有 Docker 相关命令路径已自动更新
  - 新增 `make 1panel` 命令用于打包 1Panel 应用
- **构建流程改进**：构建前自动清理 `_output` 目录，确保每次构建都是干净的环境

## [1.2.0] - 2025-01-XX

### 新增功能
- **Peer 用户绑定管理**：管理员现在可以在编辑 Peer 时修改其绑定的用户，支持将 Peer 重新分配给其他用户
  - 在编辑 Peer 对话框中添加了用户选择器（仅管理员可见）
  - 支持通过用户名选择目标用户
  - 需要 `wg_peer:update_sensitive` 权限（仅管理员）

### 修复
- **用户角色显示错误**：修复了用户列表中 admin 用户的角色显示为 "user" 的问题
  - 后端 `UserResponse` 现在正确返回 `role` 和 `status` 字段
  - 前端用户列表正确显示用户的实际角色（Admin/User）
  - 修复了单个用户查询接口也缺少角色和状态信息的问题

### 界面优化
- **用户列表显示优化**：
  - Role 列根据用户角色显示不同样式（Admin 使用默认样式，User 使用次要样式）
  - Status 列根据用户状态显示不同颜色（Active 绿色，Inactive 黄色，Deleted 红色）
  - 提升了用户列表的可读性和视觉区分度
- **Peer 编辑界面优化**：
  - 优化了编辑对话框的布局和间距
  - 改进了表单字段的提示文本和占位符

### 技术改进
- 完善了后端 API 响应结构，确保用户信息完整返回
- 优化了前端类型定义，保持前后端数据结构一致性
- 改进了错误处理和用户反馈

## [1.1.0] - 2024-XX-XX

### 首次发布
- 用户与权限管理
- 设备（Peer）管理
- IP 地址池管理
- 服务器配置自动管理

---

## 版本说明

- **新增功能**：新添加的功能特性
- **修复**：Bug 修复和问题解决
- **界面优化**：UI/UX 改进和视觉优化
- **技术改进**：代码质量、性能、架构等方面的改进
- **安全**：安全相关的更新
- **废弃**：即将移除的功能
