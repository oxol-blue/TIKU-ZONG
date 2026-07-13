# 功能补充记录：AI 计费模式

## 配置能力

AI 模型增加 `billingMode`：

- `fixed`：每次成功 AI 调用扣除 `aiChargeCount` 个 AI 配额。
- `token`：向上取整 `totalTokens / tokenUnit` 后乘以 `aiChargeCount`。
- `cost`：依据每百万 Token 成本计算实际分值，加入成本加价后，再按 `costUnitCents` 向上换算为 AI 配额。

三种模式均在 AI 成功返回且答案写入 `question_ai` 后才扣除。调用失败不会扣费。若服务商没有返回 `usage.total_tokens`，Token/成本模式会安全回退为固定次数。

在发起外部模型请求前，系统会先检查用户是否至少拥有一个可用 AI 配额；配额不足时直接返回 `NO_AI_QUOTA`，不触发第三方模型调用。模型成功后的实际扣减仍由事务性 `Consume` 处理。

## 数据库迁移

`000014_ai_billing_modes.sql` 为 `ai_models` 增加计费方式、Token 单位、每百万 Token 成本、成本加价百分比和配额单位价格字段，兼容 MySQL 5.7。

## 管理后台

AI 模型创建表单可选择计费方式；Token 与成本模式只显示相关参数，并说明向上取整规则。模型列表展示当前计费方式和固定次数/倍数。

服务商创建表单提供 DeepSeek、豆包（火山方舟）和通义千问（DashScope OpenAI 兼容模式）地址预设；管理员仍可选择自定义兼容服务地址。

## 验证

后端单元测试覆盖固定、Token、成本和缺少 Token 用量回退的配额计算，以及模型参数校验。
