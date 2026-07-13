# OCS 兼容层设计

## 1. 兼容范围

系统生成用户专属 AnswererWrapper 配置，并提供 `/api/ocs/search` 单题请求。支持 GET/POST、请求头、请求参数、`${title}`、`${question}`、`${type}`、`${options}` 占位符。

## 2. 字段解析 DSL

响应字段使用点号路径和数组下标，例如 `data.answer`、`data.items[0].value`。成功字段、成功值、题目字段和答案字段均由源配置定义。数组答案转换为 `###`。

## 3. 安全策略

只允许 HTTP/HTTPS 绝对 URL，响应体限制 4 MiB，请求超时 6 秒；不执行 OCS 配置中的任意 JavaScript handler。管理端配置的 Header 和 Token 不向普通用户展示。

## 4. 合并规则

`ANSWER_MERGE_RULE=priority` 时按源优先级故障转移；设为 `majority` 时调用全部可用源，对标准化答案投票，票数相同按优先级选择。生产环境应结合源稳定性和成本设置。
