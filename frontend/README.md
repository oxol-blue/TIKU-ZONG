<p align="center">
  <img src="https://pic4.zhimg.com/v2-702a23ebb518199355099df77a3cfe07_b.webp" width="120" height="120" alt="KOI-UI Logo" />
</p>

<h1 align="center">KOI-UI</h1>

<p align="center">
  <strong>开箱即用的 Vue 3 企业级中后台管理框架</strong><br/>
  简洁 · 美观 · 优雅 · 可扩展
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vue.js" alt="Vue" />
  <img src="https://img.shields.io/badge/TypeScript-6.x-3178C6?style=flat-square&logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Vite-8.x-646CFF?style=flat-square&logo=vite" alt="Vite" />
  <img src="https://img.shields.io/badge/Element%20Plus-2.13-409EFF?style=flat-square&logo=element" alt="Element Plus" />
  <img src="https://img.shields.io/badge/Pinia-3.x-FFD859?style=flat-square&logo=pinia" alt="Pinia" />
  <img src="https://img.shields.io/badge/UnoCSS-66.x-333333?style=flat-square&logo=unocss" alt="UnoCSS" />
</p>
<p align="center">
  <a href="https://gitee.com/KoiKite/koi-ui">
    <img src="https://gitee.com/KoiKite/koi-ui/badge/star.svg?theme=white" alt="Gitee Star" />
  </a>
  <a href="https://gitee.com/KoiKite/koi-ui">
    <img src="https://gitee.com/KoiKite/koi-ui/badge/fork.svg?theme=white" alt="Gitee Fork" />
  </a>
</p>
<p align="center">
  <a href="http://39.107.143.109/login">在线演示</a> ·
  <a href="https://gitee.com/KoiKite/koi-ui">Gitee 仓库</a> ·
  <a href="https://github.com/KoiKite/koi-ui">GitHub 仓库</a>
</p>

## 项目简介

KOI-UI 是一个开箱即用的 Vue 3 企业级中后台前端项目，内置登录、权限、动态路由、多布局、主题配置、国际化、数据看板、系统管理等常见后台能力。

项目适合作为以下场景的前端基础工程：

- 管理后台、运营平台、内容管理系统的前端脚手架
- Vue 3 + TypeScript + Element Plus 技术栈学习项目
- 前后端分离项目的后台 UI 基座
- 数据大屏、权限管理系统的二次开发模板

## 技术栈

| 分类 | 技术 |
| --- | --- |
| 核心框架 | Vue 3.5、Vue Router 5、Pinia 3 |
| 开发语言 | TypeScript 6 |
| 构建工具 | Vite 8 |
| UI 组件 | Element Plus、@element-plus/icons-vue |
| 样式方案 | SCSS、UnoCSS、CSS Variables |
| 请求通信 | Axios、SSE / EventSource |
| 图表可视化 | ECharts 6 |
| 富文本与 Markdown | WangEditor、md-editor-v3 |
| 状态持久化 | pinia-plugin-persistedstate |
| 工具增强 | VueUse、mitt、nprogress、driver.js、sortablejs、crypto-js、sm-crypto |
| 工程规范 | vue-tsc、Commitizen、cz-git |

## 预览截图

<table>
  <tr>
    <td><img src="https://pica.zhimg.com/80/v2-6888b1d2c35f2db3772223ea805fdbde_720w.webp" alt="预览 1" /></td>
    <td><img src="https://pic2.zhimg.com/80/v2-205f28eba8f1c4b76d362e5e3617deed_720w.webp" alt="预览 2" /></td>
  </tr>
  <tr>
    <td><img src="https://pica.zhimg.com/80/v2-a6d43e24a142b78c2470110b22d0befc_720w.webp" alt="KOI 1" /></td>
    <td><img src="https://pic2.zhimg.com/80/v2-7ffc8ecfaf1686a358248bc247f768af_720w.webp" alt="KOI 2" /></td>
  </tr>
</table>

## 功能特性

- 登录鉴权：支持登录页、Token 缓存、路由守卫、白名单与异常重登。
- 动态路由：支持从接口或本地 JSON 获取菜单，并自动注册动态页面。
- 权限控制：内置用户、角色、菜单、按钮权限相关页面与 `v-auth` 指令。
- 多布局：包含纵向、分栏、经典、混合、横向、双栏与移动端布局。
- 主题系统：支持亮色、暗色、灰色、色弱、主题色、头部和侧边栏反转等配置。
- 多页签：支持页签缓存、刷新、关闭、固定与 KeepAlive。
- 国际化：内置中文、英文语言包。
- 通用组件：封装 Dialog、Drawer、Upload、Photo、Excel、Search、Toolbar、RichTextView、SSE 通知等组件。
- 系统模块：包含用户、角色、菜单、部门、岗位、字典、通知、文件、图片、日志等管理页面。
- 监控模块：包含在线用户、定时任务、服务监控、Redis、缓存、黑名单等页面。
- 博客模块：包含文章、分类、标签、评论、友链、说说、文库、目录等页面。
- 数据可视化：内置工作台、分析页、控制台、Dashboard 与多套大屏素材。

## 目录结构

```text
koi-ui/
├─ plugins/                 # Vite 自定义插件
├─ public/                  # 静态资源
├─ scripts/                 # 脚本文件
├─ src/
│  ├─ api/                  # 接口定义
│  ├─ assets/               # 图片、图标、JSON 数据等资源
│  ├─ components/           # 全局业务组件
│  ├─ composables/          # 组合式函数
│  ├─ config/               # 全局配置
│  ├─ directives/           # 自定义指令
│  ├─ hooks/                # Hooks
│  ├─ languages/            # 国际化语言包
│  ├─ layouts/              # 页面布局
│  ├─ routers/              # 静态路由与动态路由
│  ├─ stores/               # Pinia 状态管理
│  ├─ styles/               # 全局样式与主题变量
│  ├─ utils/                # 工具函数
│  └─ views/                # 页面视图
├─ .env.development         # 开发环境变量
├─ .env.test                # 测试环境变量
├─ .env.production          # 生产环境变量
├─ index.html               # 应用入口模板
├─ package.json             # 项目依赖与脚本
├─ pnpm-lock.yaml           # pnpm 锁文件
├─ tsconfig.json            # TypeScript 配置
├─ unocss.config.ts         # UnoCSS 配置
└─ vite.config.ts           # Vite 配置
```

## 快速开始

### 环境要求

- Node.js >= 18
- pnpm，建议使用项目锁文件对应的包管理器

### 安装依赖

```bash
pnpm install
```

### 启动开发服务

```bash
pnpm dev
```

开发服务默认运行在：

```text
http://localhost:5730
```

Vite 配置中启用了 `open: true`，启动后会自动打开浏览器。

### 默认登录账号

开发环境默认账号配置在 `.env.development`：

```text
VITE_LOGIN_NAME = 'yuadmin'
VITE_LOGIN_PASSWORD = 'yuadmin123'
```

生产环境建议清空默认账号密码，避免将调试账号暴露给线上用户。

## 常用脚本

| 命令 | 说明 |
| --- | --- |
| `pnpm dev` | 启动开发服务 |
| `pnpm build` | 类型检查并构建默认模式产物 |
| `pnpm build:test` | 使用 test 模式构建 |
| `pnpm build:prod` | 使用 production 模式构建 |
| `pnpm type:check` | 执行 TypeScript 类型检查 |
| `pnpm release` | 使用 standard-version 生成版本 |
| `pnpm commit` | 使用 cz-git 提交并推送代码 |

> 注意：当前 `package.json` 中的 `preview` 脚本为 `npm run build:dev && vite preview`，但项目未定义 `build:dev`。如需本地预览构建产物，可先执行 `pnpm build`，再使用 `pnpm vite preview` 或调整该脚本。

## 环境变量

项目通过 `.env.development`、`.env.test`、`.env.production` 区分环境配置。所有暴露给前端代码的变量都需要以 `VITE_` 开头。

| 变量 | 说明 | 示例 |
| --- | --- | --- |
| `VITE_ENV` | 当前运行环境 | `development` |
| `VITE_APP_TITLE` | 应用标题 | `Koi-Admin` |
| `VITE_API_PREFIX` | 接口代理前缀 | `/dev-api` |
| `VITE_SERVER` | 代理目标服务地址 | `http://localhost:8088` |
| `VITE_ROUTER_MODE` | 路由模式 | `history` |
| `VITE_DROP_CONSOLE` | 构建时是否移除 `console` 和 `debugger` | `false` |
| `VITE_RESPONSE_ENCRYPT` | 是否启用响应数据加密 | `false` |
| `VITE_REQUEST_DECRYPT` | 是否启用 POST 请求数据解密 | `false` |
| `VITE_LOGIN_NAME` | 开发环境默认登录账号 | `yuadmin` |
| `VITE_LOGIN_PASSWORD` | 开发环境默认登录密码 | `yuadmin123` |

开发服务器会将 `VITE_API_PREFIX` 代理到 `VITE_SERVER`。例如 `/dev-api/user/list` 会被代理到后端服务并去除 `/dev-api` 前缀。

## 路由与权限

项目采用静态路由与动态路由结合的方式：

- 静态路由位于 `src/routers/modules/staticRouter.ts`，包含登录页、首页、错误页、工作台、大屏等基础页面。
- 动态路由位于 `src/routers/modules/dynamicRouter.ts`，登录后根据菜单权限动态注册到 `layout` 路由下。
- 本地菜单示例位于 `src/assets/json/authMenu.json`，可以作为后端菜单接口返回结构的参考。
- 权限状态由 `src/stores/modules/auth.ts`、`src/stores/modules/user.ts` 等 Pinia 模块维护。

菜单数据使用扁平结构，核心字段包括：

```json
{
  "menuId": 11,
  "menuName": "menu.system.user.name",
  "parentId": 1,
  "menuType": "2",
  "path": "/system/user",
  "name": "userPage",
  "component": "system/user/index",
  "icon": "koi-enhance-user",
  "isVisible": "1",
  "linkUrl": "",
  "isKeepAlive": "1",
  "isTag": "0",
  "tagType": "primary",
  "tagName": "New",
  "isAffix": "0",
  "redirect": ""
}
```

其中 `component` 会映射到 `src/views` 下的页面组件，例如 `system/user/index` 对应 `src/views/system/user/index.vue`。

## 主题与布局

项目内置多种后台布局，并通过全局状态和主题变量统一控制。

布局组件位于 `src/layouts`：

- `LayoutVertical`：经典左侧菜单布局
- `LayoutColumns`：分栏菜单布局
- `LayoutClassic`：顶部加侧边栏布局
- `LayoutOptimum`：混合导航布局
- `LayoutHorizontal`：顶部横向导航布局
- `LayoutDual`：双栏导航布局
- `LayoutMobile`：移动端布局

主题相关文件：

- `src/styles/theme-vars.scss`：主题变量
- `src/utils/theme.ts`：主题切换工具
- `src/utils/themeColor.ts`：主题色生成与 Element Plus 变量同步
- `src/config/presetThemeColors.ts`：预设主题色
- `src/layouts/components/ThemeConfig/index.vue`：主题配置面板

## 主要模块

| 模块 | 页面 |
| --- | --- |
| 首页 | 工作台、分析页、控制台、Dashboard |
| 系统管理 | 用户、角色、菜单、部门、岗位、字典、通知、文件、图片、消息、登录日志、操作日志 |
| 系统监控 | 在线用户、定时任务、服务监控、Redis、缓存、黑名单 |
| 博客管理 | 文章、分类、标签、评论、友链、说说、文库、目录、留言 |
| 工具模块 | 代码生成 |
| 链接模块 | 内链、iframe、Gitee、博客入口 |
| 错误页 | 403、404、500 |

## 构建部署

执行生产构建：

```bash
pnpm build:prod
```

构建产物输出到 `dist/` 目录，可部署到 Nginx、对象存储、静态站点服务或任意支持 SPA 的 Web 服务。

Nginx 示例：

```nginx
http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile        on;
    keepalive_timeout  65;

	server {
		gzip on;
		gzip_min_length  5k;
		gzip_buffers     4 16k;
		gzip_comp_level 5;
		gzip_types text/plain application/javascript application/x-javascript text/css application/xml text/javascript application/x-httpd-php image/jpeg image/gif image/png;
		gzip_vary on;
               # SSL 访问端口号为 80
		listen  80;
		listen 5730;
               # 填写绑定证书的域名,本地的是localhost
		# server_name  xxx.com localhost;
		
		# 关键：将 root 定义在这里
		root /usr/local/koi-ui/dist;
		
		# 1、version.json — 禁止缓存（版本检测必须实时）
		location = /version.json {
			add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
			add_header Pragma "no-cache";
			add_header Expires "0";
			try_files $uri =404;
		}
		# 2、index.html — 禁止缓存（入口 HTML 必须及时更新）
		location = /index.html {
			add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
			add_header Pragma "no-cache";
			add_header Expires "0";
			try_files $uri =404;
		}
		# 3、带 hash 的静态资源 — 长期缓存（Vite 构建产物文件名含 hash）
		location /assets/ {
			add_header Cache-Control "public, max-age=31536000, immutable";
			try_files $uri =404;
		}
	       
               # 前端配置
		location / {
                      # 系统资源目录，root 已从 server 继承，此处不需要再写
			# root   /usr/local/koi-ui/dist;
			# 首页位置具体文件名
			index  index.html index.htm;
			# 因为router在接收数据的时候，刷新页面出现404错误，原因是 history 造成的，配置如下解决这个问题
			try_files $uri $uri/ /index.html;
		}
	       
               # 后端配置
		location ^~/prod-api/ {
                    proxy_pass  "http://127.0.0.1:8088/";
                    proxy_headers_hash_max_size 51200;
                    proxy_headers_hash_bucket_size 6400; 
                    # 关闭重定向，让服务端看到用户的IP，而不是Nginx服务器的IP  
                    proxy_redirect off;  
                    # 设置头信息，让后端服务器能够获取到客户端的IP地址  
                    proxy_set_header X-Real-IP $remote_addr;  
                    proxy_set_header X-Forwarded-For $remote_addr;  
                    # 其他头信息，根据您的需要设置  
                    proxy_set_header Host $http_host;  
                    proxy_set_header X-Nginx-Proxy true; 
             }
		
        }
		
        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }

    }
```

## 开发建议

- 新增页面时，将页面放在 `src/views`，并确保菜单中的 `component` 字段与路径一致。
- 新增接口时，按业务模块放入 `src/api`，请求封装优先复用 `src/utils/axios.ts`。
- 新增通用组件时，放入 `src/components`，需要全局注册的组件在 `src/components/index.ts` 中导出。
- 新增主题变量时，优先维护 `theme-vars.scss` 与相关工具函数，避免页面内分散写死颜色。
- 对接后端菜单时，保持 `menuId`、`parentId`、`path`、`name`、`component`、`isVisible`、`isKeepAlive` 等字段稳定。

## 许可证

本项目基于 `LICENSE` 文件声明的许可证发布，请在使用、修改或二次分发前阅读相关条款。

## 源码与支持

| 平台 | 地址 |
| --- | --- |
| Gitee（推荐） | [https://gitee.com/KoiKite/koi-ui](https://gitee.com/KoiKite/koi-ui) |
| GitHub | [https://github.com/KoiKite/koi-ui](https://github.com/KoiKite/koi-ui) |

如果 KOI-UI 对你有帮助，欢迎在 **Gitee** 或 **GitHub** 点个 **Star**，这是对我最大的鼓励。

---

## 交流与授权

| 版本 | 技术栈 / 内容 | 参考价格 |
| --- | --- | --- |
| 前后端（基础版本） | SpringBoot 4、JDK 17、Sa-Token[无数据大屏] | 320 元 |
| 前后端（plus 版本） | SpringBoot 4、JDK 17、Sa-Token + 三个数据大屏 + 赠送博客后台 等 | 520 元 |
| 前后端（pro 版本） | SpringBoot 4、JDK 17、Sa-Token + 三个数据大屏 + AI 智能助手 + 赠送博客后台 + 前台 等 | 660 元 |
| 演示版本纯前端版本（无数据大屏） | 188 元 |
| 大屏案例 | 数据可视化大屏模板 | 150 元 / 套 |

> 加微信时请备注：**KOI-UI**。
>
> 作者闲暇时间有限，如有需要，接收个人定制咨询。

<table>
  <tr>
    <td align="center"><img src="https://gitee.com/BigCatHome/koi-photo/raw/master/photos/KOI-ADMIN/WeChat.png" alt="微信二维码" width="280" /></td>
    <td align="center"><img src="https://gitee.com/BigCatHome/koi-photo/raw/master/photos/KOI-ADMIN/WeChatPay.png" alt="微信支付" width="280" /></td>
  </tr>
</table>

## 作者

<p align="center">
  Made with ❤️ by <a href="https://gitee.com/KoiKite/koi-ui">YuXin</a>
</p>

