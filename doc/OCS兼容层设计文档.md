# OCS 兼容层设计

## 1. 兼容范围

系统生成用户专属 AnswererWrapper 配置，并提供 `/api/ocs/search` 单题请求。支持 GET/POST、请求头、请求参数、`${title}`、`${question}`、`${type}`、`${options}` 占位符。

## 2. 字段解析 DSL

响应字段使用点号路径和数组下标，例如 `data.answer`、`data.items[0].value`。成功字段、成功值、题目字段和答案字段均由源配置定义。数组答案转换为 `###`。

## 3. 安全策略

只允许 HTTP/HTTPS 绝对 URL，响应体限制 4 MiB，请求超时 6 秒；不执行 OCS 配置中的任意 JavaScript handler。管理端配置的 Header 和 Token 不向普通用户展示。

## 4. 合并规则

`ANSWER_MERGE_RULE=priority` 时按源优先级故障转移；设为 `majority` 时调用全部可用源，对标准化答案投票，票数相同按优先级选择。生产环境应结合源稳定性和成本设置。

## 5. 安全自定义字段 DSL

第三方题库源的 `data` 顶层字段除可直接填写 JSON 值外，还可使用安全转换对象：`value` 或 `template` 作为输入，按需追加 `replace`、`map`、`split`、`join`。其中 `replace` 为 `{from,to}` 数组，`map` 按字符串键映射值并可使用 `default`，`split` 和 `join` 分别用于选项文本与多值参数转换。

```json
{
  "question": { "value": "【单选题】${title}", "replace": [{"from":"单选题","to":""}] },
  "questionType": { "template": "${type}", "map": {"single": 1, "multiple": 2, "default": 0} },
  "options": { "value": "${options}", "split": "\n" }
}
```

URL、请求头和字符串参数均支持 `${title}`、`${question}`、`${type}`、`${options}`。POST 会保留 JSON 数字和数组；GET 对非标量参数进行 JSON 编码。为满足系统安全要求，任何包含 OCS 原生 `handler` JavaScript 的字段都会被拒绝，不会执行或保存为可执行逻辑。

## 6. 题库源维护

管理员可创建、查看、编辑和启用/停用第三方 OCS 题库源。编辑接口为 `PUT /api/v1/admin/ocs/sources/{id}`，状态接口为 `PATCH /api/v1/admin/ocs/sources/{id}/status`，状态值只能为 `0`（停用）或 `1`（启用）。停用后该来源不再加入本地题库后的 OCS 故障转移链路；既有配置不会被删除。
