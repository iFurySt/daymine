# 前端协作说明

Daymine 前端采用 React + TypeScript + Vite，生产构建产物由 Go 二进制内嵌并 serve。视觉方向继承 tididi：紧凑、深色优先、monospace、低圆角、面向每天反复扫描的信息 dashboard。

## 设计系统

- 深色主题为默认，保留浅色主题能力。
- 字体优先使用 `JetBrains Mono` 或等价 monospace。
- 卡片圆角不超过 8px；信息面板要密集但留足层级。
- 页面第一屏就是可用 dashboard，不做营销式 landing page。
- 内置 panel 使用统一标题、空态、错误态、刷新态和数据更新时间表达。

tididi 中可参考的前端路径：

- `web/src/index.css`：主题变量和全局字体风格。
- `web/src/pages/HomePage.tsx`：多页、多列 dashboard shell。
- `web/src/components/Component.tsx`：按 `type` 分发 widget 的早期模式。
- `web/src/components/widgets/`：Calendar、HackerNews、Weather、GitHub widget 的视觉结构。

## 必须替换的旧边界

- 不迁移 `AuthContext`、`ProtectedRoute`、登录页、邮箱验证页和 profile 权限路径。
- API client 不注入 JWT，不处理 401 登录跳转。
- Dashboard 配置不再按 user id 获取，改为 workspace 配置。

## Panel 组件边界

Panel 目标上分为四层：

- Page DSL：JSON/YAML 描述页面、列宽和 panel 布局。
- Data Source DSL：描述 panel 数据从 workspace 哪个文件、索引、目录或内置源解析。
- Renderer DSL：描述使用哪个官方 renderer、字段映射、preset 和动作。
- React Renderer Registry：前端内置 renderer 负责交互和视觉。

首批官方 renderer：

- `list`
- `card-grid`
- `timeline`
- `calendar`
- `markdown`
- `metric`
- `table`
- `schema`
- `html-template`

`feed`、`article-list`、`github-list`、`video-card`、`social-post`、`agent-runs`、`markdown-view`、`hacker-news-top` 应作为官方 preset、字段映射或受控动作配置存在，而不是长期作为硬编码 renderer。

`hacker-news-top` 读取 `index/hacker-news/top10-latest.json`，并通过 panel config 的 `task_id` 触发 `hacker-news-daily-top10`。前端只关心 panel payload，不直接知道 Codex、脚本或抓取细节。

第三方扩展先从 manifest、Markdown 模板、JSON 数据、`html-template` renderer 和 `schema` renderer 开始。`html-template` 渲染的是类似 GENUI 的块级 HTML fragment，不是完整页面；它通过 `{{ }}`、`data-for`、`data-if` 和 `dm-*` 官方标签绑定后端 resolved data。只有当内置 renderer 无法表达真实需求时，再设计受控插件 API。完整设计见 `docs/design-docs/panel-plugin-system.md`。

## 验收方式

- 开发期：Vite dev server 能连接 Go API。
- 生产期：Go 二进制能 serve 打包后的前端，刷新任意前端路由不会 404。
- UI smoke：无登录进入 dashboard，默认 panel 可见，错误态和空态不撑破布局。
- 构建检查：TypeScript build、lint 和后续浏览器 smoke 要纳入 CI。

## 当前命令

- `npm --prefix apps/web install`：安装前端依赖。
- `npm --prefix apps/web run dev`：启动 Vite dev server，API 代理到 `localhost:6345`。
- `npm --prefix apps/web run build`：执行 TypeScript 检查并把产物写入 `apps/daymine/internal/webassets/dist`。
- `make run`：构建前端并启动 Go server，启动日志会打印可直接打开的本地 URL。
