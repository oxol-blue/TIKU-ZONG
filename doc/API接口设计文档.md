# API 接口设计文档

## 1. 通用约定

JSON 响应使用 `code`、`message`、`data`。JWT 使用 `Authorization: Bearer <token>`；公开 API 搜题使用 `?key=API_KEY`。失败请求不扣次数。跨域由后端统一响应 CORS 头。

## 2. 用户接口

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/captcha
POST /api/v1/auth/refresh
GET  /api/v1/me
GET/POST /api/v1/api-key
```

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
