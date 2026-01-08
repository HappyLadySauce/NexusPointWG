# NexusPointWG v1.2.0 版本发布

## 🎉 版本概述

本次更新主要包含功能增强、Bug 修复和界面优化，提升了系统的可用性和用户体验。

---

## ✨ 新增功能

### Peer 用户绑定管理
管理员现在可以在编辑 Peer 时修改其绑定的用户，支持将 Peer 重新分配给其他用户。

**功能特点**：
- 在编辑 Peer 对话框中添加了用户选择器（仅管理员可见）
- 支持通过用户名选择目标用户
- 需要 `wg_peer:update_sensitive` 权限（仅管理员）
- 自动验证目标用户是否存在

**使用场景**：
- 员工离职后，将其设备重新分配给其他用户
- 调整设备归属关系
- 统一管理设备分配

---

## 🐛 修复

### 用户角色显示错误
修复了用户列表中 admin 用户的角色显示为 "user" 的问题。

**修复内容**：
- 后端 `UserResponse` 现在正确返回 `role` 和 `status` 字段
- 前端用户列表正确显示用户的实际角色（Admin/User）
- 修复了单个用户查询接口也缺少角色和状态信息的问题

**影响范围**：
- 用户列表页面
- 用户详情页面
- 所有涉及用户角色显示的地方

---

## 🎨 界面优化

### 🌍 国际化支持（i18n）
本次更新引入了完整的国际化支持，系统现在支持中英文双语切换。

**核心功能**：
- **双语支持**：完整支持中文和英文两种语言
- **自动语言检测**：根据浏览器语言设置自动选择默认语言
- **语言持久化**：用户选择的语言会保存到本地存储，下次访问自动应用
- **全面覆盖**：所有页面、对话框、提示信息均已国际化
  - 登录页面
  - 仪表盘
  - 设备（Peers）管理
  - IP 地址池管理
  - 用户管理
  - 设置页面

**用户体验**：
- 顶部导航栏提供便捷的语言切换器
- 语言切换即时生效，无需刷新页面
- 支持中文变体自动识别（zh-CN、zh-TW、zh-HK 统一为中文）
- 支持英文变体自动识别（en-US、en-GB 统一为英文）

### 🎯 顶部导航栏重构
重新设计了应用布局，新增固定顶部导航栏，提升用户体验。

**布局优化**：
- **顶部导航栏**：固定显示在页面顶部，包含应用标题和图标
- **用户菜单**：移至顶部导航栏右侧，包含编辑用户信息和退出登录功能
- **语言切换器**：紧凑型语言切换器，位于用户菜单左侧
- **侧边栏简化**：移除了侧边栏底部的用户信息和退出按钮，界面更简洁

**滚动优化**：
- 优化了页面滚动行为，只有设置页面支持内容滚动
- 数据展示页面（Dashboard、Peers、IP Pools、Users）使用分页而非滚动
- 避免了滚动条闪烁问题，提供更一致的视觉体验

### 📄 分页功能
为数据展示页面添加了完整的分页功能，提升大数据量场景下的使用体验。

**分页支持**：
- **Peers 页面**：支持分页浏览设备列表，每页显示 10 条数据
- **IP Pools 页面**：支持分页浏览 IP 地址池列表
- **Users 页面**：支持分页浏览用户列表
- **后端搜索**：Peers 页面的搜索功能改为后端搜索，支持分页结果

**分页特性**：
- 智能页码显示：自动计算并显示合适的页码范围
- 总数显示：显示当前数据总数
- 边界处理：删除最后一页数据时自动跳转到上一页
- 操作后刷新：创建、编辑、删除操作后自动刷新当前页

### 用户列表显示优化
- **Role 列样式优化**：
  - Admin 角色使用默认样式（更突出）
  - User 角色使用次要样式
  - 提升了角色识别的视觉区分度

- **Status 列颜色优化**：
  - Active 状态：绿色显示
  - Inactive 状态：黄色显示
  - Deleted 状态：红色显示
  - 根据状态自动调整颜色，提升可读性

### Peer 编辑界面优化
- 优化了编辑对话框的布局和间距
- 改进了表单字段的提示文本和占位符
- 优化了用户选择器的显示效果

---

## 🔧 技术改进

- **API 响应结构完善**：
  - 确保用户信息完整返回（包含 role 和 status）
  - 保持前后端数据结构一致性

- **类型定义优化**：
  - 完善了前端 TypeScript 类型定义
  - 改进了类型安全性

- **错误处理改进**：
  - 优化了用户验证的错误提示
  - 改进了用户反馈机制

---

## 📝 详细变更

### 后端变更
- `internal/pkg/types/v1/user.go`：在 `UserResponse` 中添加 `role` 和 `status` 字段
- `internal/pkg/types/v1/wg.go`：在 `UpdateWGPeerRequest` 中添加 `username` 字段
- `internal/controller/user/listUsers.go`：返回用户角色和状态信息
- `internal/controller/user/getUser.go`：返回用户角色和状态信息
- `internal/controller/wireguard/updatePeer.go`：添加用户绑定修改功能

### 前端变更
- `ui/src/app/services/api.ts`：更新 `UserResponse` 和 `UpdateWGPeerRequest` 接口
- `ui/src/app/pages/Users.tsx`：优化用户列表显示，正确显示角色和状态，添加分页功能
- `ui/src/app/pages/Peers.tsx`：添加用户选择器，支持修改 Peer 绑定用户，添加分页功能
- `ui/src/app/pages/IPPools.tsx`：添加分页功能
- `ui/src/app/components/TopBar.tsx`：新增顶部导航栏组件
- `ui/src/app/components/UserMenu.tsx`：新增用户菜单组件
- `ui/src/app/components/LanguageSwitcherCompact.tsx`：新增紧凑型语言切换器
- `ui/src/app/components/EditUserDialog.tsx`：新增编辑用户信息对话框
- `ui/src/app/context/I18nContext.tsx`：新增国际化上下文
- `ui/src/app/i18n/index.ts`：国际化配置和初始化
- `ui/src/app/i18n/locales/*/*.json`：完整的国际化翻译文件（中英文）
  - `common.json`：通用文本和按钮
  - `dashboard.json`：仪表盘页面
  - `login.json`：登录页面
  - `peers.json`：设备管理页面
  - `users.json`：用户管理页面
  - `ipPools.json`：IP 地址池页面
  - `settings.json`：设置页面
- `ui/src/app/App.tsx`：重构布局，添加顶部导航栏，优化滚动行为
- `ui/src/app/components/Sidebar.tsx`：简化侧边栏，移除用户信息显示

---

## 🔄 升级指南

### 从 v1.1.0 升级到 v1.2.0

1. **后端升级**：
   ```bash
   git pull origin main
   go mod tidy
   # 重新编译和部署
   ```

2. **前端升级**：
   ```bash
   cd ui
   pnpm install
   pnpm build
   ```

3. **数据库迁移**：
   - 无需数据库迁移
   - 现有数据完全兼容

4. **配置检查**：
   - 无需修改配置文件
   - 所有配置向后兼容

---

## 📊 统计数据

- **新增功能**：2 项（Peer 用户绑定管理、国际化支持）
- **Bug 修复**：1 项
- **界面优化**：6 项（国际化、顶部导航栏、分页功能、用户列表、Peer 编辑、滚动优化）
- **技术改进**：3 项
- **代码变更**：20+ 个文件

---

## 🙏 致谢

感谢所有贡献者和用户的支持与反馈！

---

## 📚 相关链接

- [完整变更日志](CHANGELOG.md)
- [版本管理文档](docs/VERSION_MANAGEMENT.md)
- [项目文档](README.md)

---

**Full Changelog**: https://github.com/HappyLadySauce/NexusPointWG/compare/v1.1.0...v1.2.0
