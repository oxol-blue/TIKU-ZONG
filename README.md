# TIKU-ZONG

面向学生、教师等用户的中文题库调用系统。项目提供在线搜题、API Key 调用、OCS AnswererWrapper 兼容、套餐计费和 AI 兜底，并配套用户端与管理员后台。

## 已实现能力

- 邮箱注册登录、图形验证码、登录限制、管理员 TOTP、用户/管理员角色。
- 单用户单 API Key、Key 轮换和停用、跨域 API/OCS 单题搜索。
- 题库批量导入、哈希去重、相似题匹配、选项文字答案、多选 `###` 分隔。
- 时间、次数、时间次数套餐，优惠券、订单、支付回调、退款记录、订单对账和套餐实例补偿。
- OCS 配置生成、第三方题库源、安全字段 DSL、多源优先级与多数投票。
- OpenAI 兼容 AI 服务商、多模型故障转移、缓存、队列和 AI 配额计费。
- 用户搜索/调用/反馈历史；管理员仪表盘、用户、题库、订单、支付、AI、OCS、公告、日志与系统配置管理。

## 技术栈

- 前端：Vue 3、TypeScript、Vite、Element Plus、koi-ui。
- 后端：Go、Gin、MySQL 5.7，可选 Redis。
- 部署：Nginx/宝塔、systemd、Prometheus；提供 MySQL 备份恢复脚本和部署模板。

## 本地运行

```powershell
cd backend
Copy-Item .env.example .env
go run ./cmd/migrate
go run ./cmd/server

cd ..\frontend
pnpm install
pnpm dev
```

默认前端地址为 `http://localhost:5730`，后端默认监听 `:8088`。生产环境必须替换 `.env` 中的数据库连接、JWT、加密密钥和所有服务商凭据，严禁提交真实密钥。

## 文档与验证

需求、架构、数据库、接口、部署、测试和开发过程均位于 [doc](doc/)。发布前执行：

```powershell
cd backend; go test ./...; go vet ./...
cd ..\frontend; pnpm run type:check; pnpm run build
```

真实易支付退款协议、生产部署、备份恢复演练以及真实 AI/OCS/题库数据验收仍需在目标环境完成。
