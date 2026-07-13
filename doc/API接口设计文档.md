# API 接口设计文档

## 停用 API Key

`DELETE /api/v1/api-key` 需要登录 JWT。接口将当前用户唯一 API Key 标记为已停用，API/OCS 请求立即无法再使用该 Key；历史调用日志仍保留。之后可通过创建或重新生成操作获得新的 Key。

## 管理员文件导入题库

`POST /api/v1/admin/questions/import/file` 使用 `multipart/form-data` 上传字段 `file`，仅接受 `.csv` 或 `.xlsx`，单文件不超过 10 MB、最多 1000 条数据。需要管理员 JWT；启用管理员 TOTP 时同时传递 `X-Admin-TOTP`。

首行表头支持中文或英文：`question/题目`、`type/题型`、`options/选项`、`answer/答案`、`answerRaw/原始答案`、`platform/平台`、`subject/科目`、`source/来源`、`collectedAt/采集时间`。其中题目和答案必填。选项可填写 JSON 数组，或用换行/竖线分隔的 `A. 选项文本` 格式；答案仍填写选项对应文字，多选使用 `###`。

成功响应的 `data` 包含 `total`、`valid`、`created`、`duplicates`、`invalid`、前 20 条 `preview`，及最多 100 条 `{ row, question, message }` 错误报告。有效行会导入，无效行跳过；整份文件均无效时仍返回错误报告供前端展示。

## 1. 通用约定

JSON 响应使用 `code`、`message`、`data`。JWT 使用 `Authorization: Bearer <token>`；公开 API 搜题使用 `?key=API_KEY`。失败请求不扣次数。跨域由后端统一响应 CORS 头。

## 2. 用户接口

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/captcha
POST /api/v1/auth/refresh
GET  /api/v1/me
POST /api/v1/password/change
GET/POST /api/v1/api-key
```

`POST /api/v1/password/change` 需要登录态，提交 `currentPassword` 与 `newPassword`。新密码长度为 8–72 个字符且不能与当前密码相同；服务端使用 bcrypt 保存新密码，并撤销该用户全部刷新令牌。已签发的访问令牌仅在其正常短期过期前仍可使用。

## 3. 搜题接口

```text
GET /api/v1/search?q=题目&type=single&options=A.%20选项&key=API_KEY
GET /api/ocs/config?key=API_KEY
GET /api/ocs/search?key=API_KEY&q=题目
```

普通搜题返回题目、文字答案、题型、`is_ai`、`similarity`、耗时和来源；OCS 成功格式为 `{code:1,q,data}`。

## 4. 管理接口

管理接口包括用户、题库、套餐、优惠券、订单、调用日志、AI provider/model/answer、OCS source 和支付网关接口。敏感管理接口要求管理员身份；配置 `ADMIN_TOTP_SECRET` 后还要求 `X-Admin-TOTP`。

完整路由以 `backend/internal/httpapi/router.go` 为准。
### 管理员查询退款明细

`GET /api/v1/admin/orders/{orderNo}/refunds` 返回该订单的退款号、金额、原因、渠道状态和创建时间。退款提交支持可选 `refundNo`，相同退款号重复提交时返回原订单状态，不会重复累计退款金额。

### 服务探针

- `GET /healthz`：进程存活探针，不依赖数据库。
- `GET /readyz`：服务就绪探针，检查 MySQL 连接；数据库未配置或不可用时返回 `503`。
- `GET /metrics`：Prometheus 文本格式的非敏感运行指标。

### 管理统计

`GET /api/v1/admin/dashboard`（管理员）返回用户数、付费用户数、已支付订单/金额、API 调用量、在线搜索量、成功调用量、本地/OCS/AI 命中量、AI Token 累计、套餐消耗额度、错误率和平均响应耗时。新写入的调用日志带有 `sourceKind`（`local`、`ocs`、`ai`）以支持该统计。

退款管理页使用 `GET /api/v1/admin/orders/{orderNo}/refunds` 查看退款记录，并在退款提交 JSON 中传递唯一 `refundNo`。

`GET /api/v1/admin/orders/reconciliation` 执行只读订单对账，检查已支付订单缺少套餐实例、过期未关闭订单以及退款累计不一致等异常。

`POST /api/v1/admin/orders/repair-package-instances` 为已支付或部分退款、但缺少套餐实例关联的历史订单补偿套餐实例。补偿始终以原支付时间作为套餐开始时间，沿用订单对应套餐的有效期、普通次数和 AI 次数；全额退款订单不会补偿。事务锁定订单并要求 `package_instance_id = 0`，重复调用安全且不会重复发放。

`GET /api/v1/calls/my`（登录用户）返回当前用户最近调用记录，不返回题目明文，仅返回题目哈希、接口、来源、耗时和状态。

`GET /api/v1/search-history/my`（登录用户）返回当前用户成功搜索的题目、文字答案、题型、来源、AI 标记、耗时和时间。支持 `page`、`pageSize` 分页以及 `isAi=true|false` 筛选；历史仅按当前 JWT 用户 ID 查询，API 调用日志中的失败请求不会写入该记录。

`GET /api/v1/feedback/my`（登录用户）返回当前用户提交的反馈记录，仅返回题目哈希，不返回题目原文。

`GET /api/v1/admin/feedback`（管理员）支持 `search`、`type`、`page` 和 `pageSize` 筛选反馈记录。

### 公告与支付配置

```text
GET   /api/v1/announcements
GET   /api/v1/admin/announcements
POST  /api/v1/admin/announcements
PUT   /api/v1/admin/announcements/{id}
PATCH /api/v1/admin/announcements/{id}/status
GET   /api/v1/admin/payment/gateways?provider=epay
```

公告管理支持发布、编辑、置顶和上下架；支付网关查询只返回基础配置和 `keyConfigured`，不返回支付密钥明文。

### 套餐与优惠券维护

```text
PUT   /api/v1/admin/packages/{id}
PATCH /api/v1/admin/packages/{id}/status
POST  /api/v1/admin/packages/{id}/grant/{userId}
PATCH /api/v1/admin/coupons/{id}/status
```

套餐发放接口会直接为指定用户创建套餐实例，不会创建支付订单；只允许发放当前已上架套餐。

### AI 模型维护

```text
GET   /api/v1/admin/ai/models
PUT   /api/v1/admin/ai/models/{id}
PATCH /api/v1/admin/ai/models/{id}/status
```

模型可修改服务商、名称、优先级、超时与计费参数。调用链按优先级从小到大依次尝试，只有启用的模型会进入故障转移；列表不会返回服务商 API Key 明文。

### OCS 第三方题库源

`POST /api/v1/admin/ocs/sources` 接收名称、主页、URL、GET/POST 方法、请求头、`data`、优先级、启用状态及响应字段路径。`data` 的顶层字段可使用安全 DSL：`value/template`、`replace`、`map`、`split`、`join`；字符串值和 URL 支持 `${title}`、`${question}`、`${type}`、`${options}`。包含 OCS 原生 JavaScript `handler` 的字段会返回参数错误，服务端不执行任意脚本。

```text
PUT   /api/v1/admin/ocs/sources/{id}
PATCH /api/v1/admin/ocs/sources/{id}/status  { "status": 0|1 }
```

编辑请求使用与创建相同的完整题库源载荷。不存在的来源返回 `404 OCS_SOURCE_NOT_FOUND`；状态字段缺失或不是 `0`/`1` 返回 `400`。

### 管理员操作日志

```text
GET /api/v1/admin/audit-logs?page=1&pageSize=20&search=keyword
```

仅管理员可访问。日志记录成功的 `POST`、`PUT`、`PATCH`、`DELETE` 管理操作，返回操作者、动作、资源路径、IP、HTTP 状态和时间；不会保存请求正文，因此不会写入密码、API Key、AI 密钥或支付密钥。

套餐编辑会校验套餐类型、有效期、次数和价格；停用优惠券不会影响已经创建的订单。
