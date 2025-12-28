# 前后端接口对接规范（NexusPointWG）

本文档用于规范 NexusPointWG 项目**前端（ui）**与**后端（api）**的接口对接方式，确保：
- **统一响应结构**（code/message/details）
- **统一错误码语义**（前端以 code 为准）
- **校验错误可国际化**（details 使用 token）
- **鉴权一致**（JWT Bearer）
- **接口新增/变更可追溯**（Swagger + 规范约束）

> 适用范围：所有 `BasePath=/api/v1` 的 JSON API。

---

## 1. 基本约定

### 1.1 BasePath 与版本
- **BasePath**：`/api/v1`
- **版本策略**：不破坏兼容的小改动不升版本；破坏性变更必须新开 `/api/v2`。

### 1.2 Content-Type
- 请求：`Content-Type: application/json`
- 响应：`application/json; charset=utf-8`

### 1.3 鉴权（JWT Bearer）
- Header：`Authorization: Bearer <token>`
- 前端规则：
  - 无 token：仅允许访问登录/注册页。
  - token 过期/无效：清 token 并跳转登录页。

---

## 2. 统一响应结构（后端输出）

后端统一使用 `pkg/core` 的响应结构：
- `core.SuccessResponse`
- `core.ErrResponse`

### 2.1 成功响应（无 data）

```json
{
  "code": 100001,
  "message": "OK"
}
```

### 2.2 错误响应（统一格式）

```json
{
  "code": 100003,
  "message": "Error occurred while binding the request body to the struct",
  "reference": "",
  "details": {
    "email": "validation.emaildomain|domains=qq.com,163.com,gmail.com,outlook.com",
    "username": "validation.urlsafe"
  }
}
```

字段说明：
- **code**：业务错误码（前端必须优先使用 code 进行提示映射）
- **message**：后端注册的外部消息（仅作为兜底，不建议直接展示给终端用户）
- **reference**：可选的参考链接（目前多数为空）
- **details**：可选字段级错误（注册/更新/改密等校验失败时出现）

---

## 3. 校验错误 details 规范（i18n 关键）

### 3.1 details token 格式

`details[field]` 统一输出 token（后端实现位于 `pkg/core/core.go::FormatValidationError`）：

- **格式**：`validation.<tag>|k=v|k2=v2`
- 示例：
  - `validation.required`
  - `validation.min|min=8`
  - `validation.emaildomain|domains=qq.com,163.com,gmail.com,outlook.com`

> 前端必须解析 token，并用 i18n 渲染；不要把后端英文句子直接展示给用户。

### 3.2 tag 列表（当前实现）
- `validation.required`
- `validation.min|min=<n>`
- `validation.max|max=<n>`
- `validation.len|len=<n>`
- `validation.email`
- `validation.oneof|values=a,b,c`
- `validation.urlsafe`
- `validation.nochinese`
- `validation.emaildomain|domains=...`
- `validation.invalid|tag=<tag>`

---

## 4. 错误码规范（前端提示以 code 为准）

错误码定义位于：
- 通用/鉴权：`internal/pkg/code/base.go`
- 业务（用户相关）：`internal/pkg/code/server.go`
- 注册消息：`internal/pkg/code/register.go`

### 4.1 通用/鉴权（摘自 `internal/pkg/code/base.go`）
- `100001`：成功（OK）
- `100003`：绑定/校验失败（Validation failed / Bind error），`details` 中包含字段级 token
- `100201`：加密错误（ErrEncrypt）
- `100202`：签名无效（ErrSignatureInvalid）
- `100203`：token 过期（ErrExpired）
- `100204`：Authorization header 无效（ErrInvalidAuthHeader）
- `100205`：Authorization header 缺失（ErrMissingHeader）
- `100206`：密码错误（ErrPasswordIncorrect）
- `100207`：无权限（ErrPermissionDenied）

### 4.2 用户相关（摘自 `internal/pkg/code/server.go`）
- `110001`：用户名已存在（ErrUserAlreadyExist）
- `110002`：邮箱已存在（ErrEmailAlreadyExist）
- `110003`：用户不存在（ErrUserNotFound）
- `110004`：用户未激活（ErrUserNotActive）

---

## 5. 登录失败提示策略（安全规范）

为避免用户名枚举攻击，登录失败不区分“用户不存在”或“密码错误”。

### 5.1 后端建议
- 后端对用户名不存在与密码错误可以返回同一业务码（当前实现倾向如此）。

### 5.2 前端必须遵循
- 对 `100206`（以及若出现的 `110003`），**统一展示**：
  - zh-CN：`用户名或密码错误`
  - en-US：`Incorrect username or password`

> 前端不应直接展示后端 `message`（如 `Password was incorrect`）。

---

## 6. 现有接口清单（基于 Swagger）

Swagger 文件：`api/swagger/docs/swagger.json`

### 6.1 Auth

#### POST `/api/v1/login`（用户登录）
- **Request**：`v1.LoginRequest`（见 Swagger definitions）
- **200 Response**：`v1.LoginResponse`（包含 `token`）
- **401/403/400**：`core.ErrResponse`

示例：

**Request**
```json
{ "username": "test", "password": "12345678" }
```

**200 Response**
```json
{ "token": "<jwt>" }
```

### 6.2 Users

#### POST `/api/v1/users`（用户注册）
- **Request**：`v1.CreateUserRequest`（定义：`internal/pkg/types/v1/user.go`）
- **200 Response**：`core.SuccessResponse`
- **400**：
  - 校验失败：`code=100003`，`details` 返回 `validation.*` token
  - 冲突：`code=110001/110002`

示例：

**Request**
```json
{
  "username": "test",
  "nickname": "test",
  "avatar": "https://example.com/a.png",
  "email": "test@163.com",
  "password": "12345678"
}
```

#### GET `/api/v1/users/{username}`（获取用户信息）
- **200 Response**：`v1.UserResponse`

#### PUT `/api/v1/users/{username}`（更新用户）
- **Request**：`v1.UpdateUserRequest`
- 普通用户：仅允许更新本人，且仅允许更新 `username/nickname/avatar/email`
- 管理员：可更新更多字段（`password/status/role` 等）

#### DELETE `/api/v1/users/{username}`（删除用户）
- 普通用户：仅允许对本人做注销/软删除（具体后端实现为准）
- 管理员：可删除任意用户

#### POST `/api/v1/users/{username}/password`（修改密码）
- **Request**：`v1.ChangePwdRequest`

---

## 7. 前端处理规范（必须遵循）

### 7.1 错误展示优先级
1. **code → i18n 文案**（最高优先级）
2. 若存在 `details`：逐字段解析 token → `t(key, params)` 渲染
3. 最后兜底：`message` 或通用 `common.unknownError`

### 7.2 注册页（Form）
- 400 校验失败：展示字段级错误（可 toast + 表单红字；建议后续升级为 `Form.setFields`）
- 110001/110002：直接提示已存在（i18n）

### 7.3 登录页（安全策略）
- 对 `100206` 统一提示 `error.authFailed`
- 不展示“用户不存在”等可被枚举的信息

---

## 8. 后端开发规范（新增接口必须包含）

### 8.1 Swagger 注解
新增/变更接口必须补齐 swagger 注解（示例见 `internal/controller/**`）：
- `@Summary @Description @Tags @Accept @Produce`
- `@Param`（path/body）
- `@Success @Failure`
- `@Router`

### 8.2 统一返回
- 绑定/校验失败：使用 `core.WriteResponseBindErr`
- 业务错误：使用 `core.WriteResponse`
- 校验错误细节：由 `FormatValidationError` 生成 token（不要在 controller 手写英文句子）

---

## 9. 变更流程（建议）

### 9.1 新增/变更接口
- 后端先更新 swagger + types（request/response）
- 若新增错误码：必须在 `internal/pkg/code/*` 中注册，并在本规范“错误码表”补充说明
- 前端同步维护：
  - 错误码 → i18n key 映射
  - i18n 文案（zh-CN/en-US）

### 9.2 PR Checklist（建议）
- [ ] 响应结构符合 `core.SuccessResponse/core.ErrResponse`
- [ ] 错误码稳定、语义清晰、前端可 i18n
- [ ] 校验错误 details 使用 `validation.*` token
- [ ] Swagger 文档已更新

---

## 10. 字段定义与校验规则（后端 binding tags）

本章节用于前端实现表单校验与错误提示展示；规则来源于后端 struct 的 `binding` tag：
- `internal/pkg/types/v1/user.go`
- `api/swagger/docs/swagger.json`（definitions）

> 说明：后端校验失败会返回 `code=100003`，并在 `details` 中返回 `validation.*` token（见第 3 章）。

### 10.1 登录 `v1.LoginRequest`

| 字段 | 类型 | 必填 | 约束/说明 |
|---|---|---:|---|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

### 10.2 注册 `v1.CreateUserRequest`

| 字段 | 类型 | 必填 | 约束（binding） | 说明 |
|---|---|---:|---|---|
| username | string | 是 | `required,min=3,max=32,urlsafe,nochinese` | 3-32 字符；仅字母/数字/下划线/连字符；不能包含中文 |
| nickname | string | 否 | `omitempty,min=3,max=32` | 3-32 字符；不传则默认用 username |
| avatar | string(URL) | 否 | `omitempty,url,max=255` | 头像 URL |
| email | string | 是 | `required,email,emaildomain,max=255` | 邮箱格式；域名必须在允许列表内 |
| password | string | 是 | `required,min=8,max=32` | 8-32 字符 |

### 10.3 更新用户 `v1.UpdateUserRequest`（部分更新）

| 字段 | 类型 | 必填 | 约束（binding） | 说明 |
|---|---|---:|---|---|
| username | string | 否 | `omitempty,min=3,max=32,urlsafe,nochinese` | 同注册规则 |
| nickname | string | 否 | `omitempty,min=3,max=32` | 同注册规则 |
| avatar | string(URL) | 否 | `omitempty,url,max=255` | 同注册规则 |
| email | string | 否 | `omitempty,email,emaildomain,max=255` | 同注册规则 |
| password | string | 否 | `omitempty,min=8,max=32` | 仅管理员允许（业务层控制） |
| status | string | 否 | `omitempty,oneof=active inactive deleted` | 仅管理员允许（业务层控制） |
| role | string | 否 | `omitempty,oneof=user admin` | 仅管理员允许（业务层控制） |

### 10.4 修改密码 `v1.ChangePwdRequest`

| 字段 | 类型 | 必填 | 约束（binding） | 说明 |
|---|---|---:|---|---|
| old_password | string | 是 | `required,min=8,max=32` | 旧密码 |
| new_password | string | 是 | `required,min=8,max=32` | 新密码 |

---

## 11. 错误码 → 前端 i18n 映射表（建议实现）

### 11.1 映射原则
1. 前端提示优先使用 `code`（而不是后端英文 `message`）
2. 登录失败遵循安全策略：不区分用户不存在/密码错误 → 统一提示 `error.authFailed`
3. 表单校验失败：使用 `details` 中的 `validation.*` token（第 3 章）渲染字段级错误

### 11.2 建议映射表（与 `ui/src/utils/request.ts` 对齐）

| code | 后端含义 | HTTP | 前端 i18n key | 备注 |
|---:|---|---:|---|---|
| 100201 | ErrEncrypt | 401 | `error.encrypt` | |
| 100202 | ErrSignatureInvalid | 401 | `error.tokenInvalid` | |
| 100203 | ErrExpired | 401 | `error.tokenExpired` | |
| 100204 | ErrInvalidAuthHeader | 401 | `error.tokenInvalid` | |
| 100205 | ErrMissingHeader | 401 | `error.tokenInvalid` | |
| 100206 | ErrPasswordIncorrect | 401 | `error.authFailed` | **登录失败统一提示** |
| 100207 | ErrPermissionDenied | 403 | `error.permissionDenied` | |
| 110001 | ErrUserAlreadyExist | 400 | `error.userAlreadyExist` | 注册用户名重复 |
| 110002 | ErrEmailAlreadyExist | 400 | `error.emailAlreadyExist` | 注册邮箱重复 |
| 110003 | ErrUserNotFound | 404/401 | `error.authFailed` 或 `error.userNotFound` | **登录页建议统一 `authFailed`**；其它业务可用 `userNotFound` |
| 110004 | ErrUserNotActive | 403 | `error.userNotActive` | |

### 11.3 validation token → i18n key
后端 `details[field]` 为 token：`validation.<tag>|...`，前端解析后调用 `t('validation.<tag>', params)`。



