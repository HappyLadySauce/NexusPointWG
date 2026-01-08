# 版本管理文档

本文档用于记录版本更新流程和变更追踪方法，方便在发布新版本时快速生成 Release Notes。

## 版本号规范

项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/) (Semantic Versioning)：

- **主版本号 (MAJOR)**：不兼容的 API 修改
- **次版本号 (MINOR)**：向下兼容的功能性新增
- **修订号 (PATCH)**：向下兼容的问题修正

格式：`MAJOR.MINOR.PATCH`，例如：`1.2.0`

## 版本更新流程

### 1. 开发阶段

在开发新功能或修复 Bug 时，在对应的变更类型下记录变更：

```markdown
## [未发布]

### 新增功能
- [ ] 功能描述

### 修复
- [ ] Bug 描述

### 界面优化
- [ ] 优化描述
```

### 2. 发布前准备

1. **更新版本号**
   - 更新 `pkg/environment/version.go` 中的版本号
   - 更新 `CHANGELOG.md`，将 `[未发布]` 改为具体版本号和日期

2. **检查变更清单**
   - 确认所有变更都已记录
   - 检查是否有遗漏的功能或修复
   - 验证变更分类是否正确

3. **生成 Release Notes**
   - 从 `CHANGELOG.md` 复制对应版本的变更内容
   - 根据 GitHub Release 格式调整格式
   - 添加适当的 emoji 和格式化

### 3. 发布后

1. **创建 Git Tag**
   ```bash
   git tag -a v1.2.0 -m "Release v1.2.0"
   git push origin v1.2.0
   ```

2. **创建 GitHub Release**
   - 使用 `CHANGELOG.md` 中的内容作为 Release Notes
   - 添加适当的标签和分类

## 变更分类指南

### 新增功能 (Features)
- 新添加的功能特性
- 新的 API 端点
- 新的配置选项
- 新的用户界面组件

**示例**：
- 新增 Peer 用户绑定管理功能
- 新增批量操作功能
- 新增数据导出功能

### 修复 (Bug Fixes)
- Bug 修复
- 错误处理改进
- 数据一致性问题修复
- 安全漏洞修复

**示例**：
- 修复用户角色显示错误
- 修复 IP 地址分配冲突问题
- 修复权限检查逻辑错误

### 界面优化 (UI/UX Improvements)
- 界面布局调整
- 样式美化
- 交互体验改进
- 响应式设计优化
- 国际化文本更新

**示例**：
- 优化用户列表显示样式
- 改进表单验证提示
- 优化移动端显示效果

### 技术改进 (Technical Improvements)
- 代码重构
- 性能优化
- 架构改进
- 依赖更新
- 文档完善

**示例**：
- 优化数据库查询性能
- 重构权限检查逻辑
- 更新依赖包版本
- 完善 API 文档

### 安全 (Security)
- 安全漏洞修复
- 权限控制加强
- 数据加密改进

**示例**：
- 修复 SQL 注入漏洞
- 加强密码加密强度
- 改进 Token 验证机制

### 废弃 (Deprecations)
- 即将移除的功能
- 废弃的 API
- 迁移指南

**示例**：
- 废弃旧的用户创建接口
- 计划移除某个配置选项

## 版本变更记录模板

在 `CHANGELOG.md` 中使用以下模板：

```markdown
## [版本号] - YYYY-MM-DD

### 新增功能
- 功能描述 1
- 功能描述 2

### 修复
- Bug 描述 1
- Bug 描述 2

### 界面优化
- 优化描述 1
- 优化描述 2

### 技术改进
- 改进描述 1
- 改进描述 2
```

## GitHub Release Notes 格式

发布到 GitHub 时，可以使用以下格式：

```markdown
## 🎉 NexusPointWG v1.2.0 版本发布

### ✨ 新增功能
- **Peer 用户绑定管理**：管理员现在可以在编辑 Peer 时修改其绑定的用户

### 🐛 修复
- 修复了用户列表中 admin 用户的角色显示错误

### 🎨 界面优化
- 优化了用户列表的显示样式，根据角色和状态显示不同颜色
- 改进了 Peer 编辑界面的布局

### 🔧 技术改进
- 完善了后端 API 响应结构
- 优化了前端类型定义
```

## 版本历史追踪

### v1.2.0 (当前版本)
- **主要变更**：Peer 用户绑定管理、用户角色显示修复、界面优化
- **发布日期**：2025-01-XX
- **变更类型**：功能新增、Bug 修复、界面优化

### v1.1.0
- **主要变更**：首次发布
- **发布日期**：2024-XX-XX
- **变更类型**：初始版本

## 快速检查清单

发布新版本前，请确认：

- [ ] 版本号已更新（`pkg/environment/version.go`）
- [ ] `CHANGELOG.md` 已更新
- [ ] 所有变更都已分类记录
- [ ] 已测试所有新功能和修复
- [ ] 文档已更新（如需要）
- [ ] Git tag 已创建
- [ ] GitHub Release 已创建

## 注意事项

1. **保持变更记录及时更新**：在开发过程中就记录变更，不要等到发布前才补充
2. **变更描述要清晰**：使用简洁明了的语言描述变更，避免技术术语过多
3. **分类要准确**：确保每个变更都放在正确的分类下
4. **版本号要一致**：确保代码、文档、Release 中的版本号一致
5. **日期格式统一**：使用 `YYYY-MM-DD` 格式

## 参考资源

- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)
