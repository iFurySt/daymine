# Panel Plugin System 设计

## 当前状态

当前 MVP 已经把页面配置、panel 数据和运行产物放到 workspace：

- 页面和 panel manifest：`config/daymine.json`
- 通用 panel 数据：`index/panels.json`
- Agent run：`index/runs.jsonl`
- Markdown 内容：`notes/**/*.md`
- HTML fragment template：`config/panels/*.template.html`
- Panel-local CSS：`config/panels/*.panel.css`

当前已跑通一个外置 `html-template` panel：后端从 workspace 读取 template/style/data source，前端运行时解析 HTML fragment，并把 `dm-*` 标签映射到内置 React 组件。启动时会给缺少该 panel 的既有 workspace 追加默认 `external-signal` 配置和示例数据，不覆盖已有 panel。仍未完全解耦的部分是：内置 panel 还在 `apps/web/src/App.tsx` 里用 `switch(panel.type)` 选择 React renderer，后端也仍对 `calendar`、`agent-runs`、`markdown-view` 做特判。下一步要把这些收敛到统一 renderer registry。

## 设计目标

- 一个二进制启动后同时拥有前端和后端。
- AI 可以通过修改 workspace 文件管理页面、布局、panel、数据源和 renderer 配置。
- 官方 panel 内置在二进制里，稳定、可测、可复用。
- 用户和社区可以按官方规范创建自定义 panel，不需要重新编译二进制即可覆盖大部分信息展示。
- 只有确实需要复杂交互时，才进入更高风险的扩展机制。

## 分层模型

Panel 系统分五层：

1. Page DSL：描述有哪些页面、列宽、布局和 panel 实例。
2. Data Source DSL：描述 panel 数据从哪里来、如何查询、如何刷新。
3. Template DSL：描述 HTML fragment、字段绑定、循环、条件和动作。
4. Renderer DSL：描述用哪个官方 renderer、HTML template renderer 或 schema renderer。
5. Renderer Runtime：前端内置官方 renderer registry，后端提供规范化数据 API。

```text
workspace/config/pages/*.json
  -> page layout
workspace/config/panels/*.json
  -> panel instance + data source + renderer config
workspace/index/*.json / notes/**/*.md / artifacts/**
  -> data
Go API
  -> validates manifests, resolves data, returns normalized panel payload
React runtime
  -> renderer registry + HTML fragment template runtime + generic schema renderer
```

## 推荐 DSL

### 页面

```json
{
  "schema_version": "daymine.page.v1",
  "name": "home",
  "title": "Home",
  "columns": [
    { "id": "left", "width": 1, "panels": ["today", "rss"] },
    { "id": "center", "width": 2, "panels": ["articles", "github"] },
    { "id": "right", "width": 1, "panels": ["runs", "daily-note"] }
  ]
}
```

### Panel

```json
{
  "schema_version": "daymine.panel.v1",
  "id": "rss",
  "title": "RSS Feed",
  "renderer": {
    "type": "list",
    "variant": "feed",
    "fields": {
      "title": "title",
      "summary": "summary",
      "meta": "source",
      "href": "url",
      "time": "published_at"
    }
  },
  "data": {
    "type": "json",
    "path": "index/panels.json",
    "selector": "$.feed",
    "limit": 20,
    "sort": [{ "field": "published_at", "direction": "desc" }]
  },
  "refresh": "15m",
  "actions": [
    {
      "id": "summarize",
      "label": "Summarize",
      "kind": "agent-run",
      "provider": "codex-cli",
      "query_template": "Summarize RSS item: {{item.title}} {{item.url}}"
    }
  ]
}
```

### HTML Fragment Panel

HTML template panel 是推荐的动态自定义能力。它的理念接近 Go `html/template`：数据和渲染分离，模板持久化在 workspace，运行时把 resolved data 绑定进去。它也接近 Claude/GENUI 里的块级 HTML：LLM 产出的是一个 fragment，例如 `<div>...</div>`，不是完整的 `<html>` 页面。

```json
{
  "schema_version": "daymine.panel.v1",
  "id": "rss-html",
  "title": "RSS Feed",
  "renderer": {
    "type": "html-template",
    "template_path": "config/panels/rss.template.html",
    "style_path": "config/panels/rss.panel.css"
  },
  "data": {
    "type": "json",
    "path": "index/panels.json",
    "selector": "$.feed",
    "as": "items",
    "limit": 20
  }
}
```

模板示例：

```html
<dm-list>
  <dm-item data-for="item in items">
    <dm-link href="{{ item.url }}">{{ item.title }}</dm-link>
    <dm-text tone="muted" max-lines="3">{{ item.summary }}</dm-text>
    <dm-meta>{{ item.source }} · {{ item.published_at }}</dm-meta>
  </dm-item>
</dm-list>
```

这里的 `dm-*` 标签不是浏览器原生自定义元素，也不是任意第三方 Web Component；它们是 Daymine runtime 映射到官方 React 组件的受控标签。这样 AI 可以写接近 HTML 的 UI，前端仍然能保持统一样式、空态、错误态、动作和安全边界。

## HTML Template Contract

HTML template 不是完整页面，而是 panel body fragment。

允许：

- 安全 HTML 结构：`div`、`section`、`header`、`ul`、`li`、`p`、`span`、`a`、`img` 等白名单标签。
- Daymine 官方标签：`dm-list`、`dm-item`、`dm-card`、`dm-title`、`dm-text`、`dm-meta`、`dm-badge`、`dm-link`、`dm-time`、`dm-markdown`、`dm-table`。
- 数据绑定：`{{ path.to.value }}`。
- 循环：`data-for="item in items"`。
- 条件：`data-if="item.url"`。
- 官方动作：`data-action="agent-run:summarize"`。
- 块级样式：受限 CSS custom properties、class、部分 layout/style 属性。

禁止：

- `<script>`。
- `on*` inline event，例如 `onclick`。
- `javascript:`、`data:` 等危险 URL。
- 任意外链脚本、iframe、object/embed。
- 直接访问文件系统、环境变量或浏览器危险 API。

绑定表达式第一阶段只支持 path，不支持任意 JS 表达式。后续可以增加受控 formatter，例如：

```html
<dm-time value="{{ item.published_at }}" format="relative"></dm-time>
<dm-text>{{ item.summary | truncate:160 }}</dm-text>
```

CSS 的建议边界：

- 可以允许 panel-local CSS fragment，但必须做 selector scope，例如自动包到 `[data-panel-id="rss-html"]`。
- 允许颜色、间距、display、grid/flex、字体粗细等低风险属性。
- 禁止 `position: fixed`、全局 selector、外部 `@import`、动画滥用和遮挡页面的样式。
- 更推荐使用 `dm-*` 组件属性表达样式，而不是生成大量自由 CSS。

## 数据绑定模型

后端负责把 Data Source DSL 解析成统一 context：

```json
{
  "panel": {
    "id": "rss-html",
    "title": "RSS Feed"
  },
  "items": [
    {
      "title": "Example",
      "summary": "Short summary",
      "url": "https://example.com",
      "source": "RSS",
      "published_at": "2026-04-22T10:00:00Z"
    }
  ]
}
```

模板只能访问这个 context。这样 AI 需要做的是同时维护两份 contract：

- `data.as` 或 selector 产出的变量名，例如 `items`。
- HTML fragment 中使用的路径，例如 `item.title`、`item.url`。

这解决了 “Panel HTML 如何对接数据” 的核心问题：不是 HTML 自己去请求数据，而是 panel manifest 声明 data source，后端 resolve 成 context，前端 template runtime 只做绑定和渲染。

## 官方 Renderer

第一阶段官方 renderer 应覆盖信息密度最高的形态：

| Renderer | 用途 | 可配置点 |
| --- | --- | --- |
| `list` | RSS、文章、GitHub、社媒条目 | 字段映射、meta、时间、链接、标签、limit |
| `card-grid` | 视频、项目、图片、资源卡 | 标题、封面、描述、badge、列数 |
| `timeline` | run history、事件流、变更记录 | 时间字段、状态字段、分组 |
| `calendar` | 日历和每日入口 | 日期字段、事件字段 |
| `markdown` | 渲染 Markdown 文件或片段 | path、heading、max height |
| `metric` | 小型数字/状态摘要 | value、label、trend、unit |
| `table` | 结构化列表、任务、库存 | columns、sort、compact |
| `schema` | 通用 fallback | JSON schema 或 UI schema |
| `html-template` | AI/用户持久化的 HTML fragment panel | template、style、binding、actions |

关键点：`feed`、`github-list`、`article-list` 不应该长期作为硬编码 renderer，而应该是 `list` renderer 的官方 preset。

## 自定义 Panel 的层级

### Level 1：Preset

用户只选择官方 preset，例如 `preset: "rss-feed"`。AI 主要填数据源和字段。

适合大多数社区贡献和普通用户。

### Level 2：HTML Template Renderer

用户或 AI 写 HTML fragment template，使用 `dm-*` 官方标签和受控数据绑定。

这层是推荐的动态自定义方式，适合 GENUI 风格的持久化 panel。它比 schema renderer 更接近人和 LLM 都熟悉的 HTML，但仍不执行任意 JS。

### Level 3：Schema Renderer

用户写 `renderer.type = "schema"`，用受控 UI schema 描述文本、列表、badge、链接、时间、图片等元素。

这层不执行任意 JS，适合 AI 自动生成。

示例：

```json
{
  "renderer": {
    "type": "schema",
    "layout": "stack",
    "children": [
      { "kind": "text", "field": "title", "style": "title" },
      { "kind": "text", "field": "summary", "style": "muted", "max_lines": 3 },
      { "kind": "link", "field": "url", "label": "Open" }
    ]
  }
}
```

### Level 4：Static Web Component

后续可以允许 workspace 中的 `extensions/panels/<name>/` 提供前端 bundle，但默认禁用或明确标记为 trusted。

这层才允许复杂交互和第三方代码，必须有安全边界：

- 只从本地 workspace 加载。
- 显式开启 trusted extensions。
- 不给任意文件系统权限，只通过受控 API 读 panel payload。
- 记录扩展来源和版本。

## Data Source DSL

第一阶段支持：

- `json`：读取 workspace 内 JSON 文件，支持简单 selector。
- `jsonl`：读取 append-only log，支持 limit 和倒序。
- `markdown`：读取 Markdown 文件。
- `directory`：列目录，适合 artifacts、attachments、daily notes。
- `agent-query`：触发或读取某个 Agent run 的产物。
- `builtin`：例如 calendar、workspace status。

后续再加：

- `rss`：后端抓取并缓存。
- `github`：读取 token 或公开 API，产出标准 JSON。
- `sqlite`：当可选索引层落地后支持 SQL/FTS。

## 文件布局

建议把默认 `config/daymine.json` 拆小：

```text
~/.daymine/
  config/
    daymine.json
    pages/
      home.json
      research.json
    panels/
      today.json
      rss.json
      articles.json
      github.json
      runs.json
    presets/
      rss-feed.json
      github-repos.json
  extensions/
    panels/
  index/
    panels.json
    runs.jsonl
```

`config/daymine.json` 只负责应用级配置和页面入口：

```json
{
  "schema_version": "daymine.app.v1",
  "pages": ["config/pages/home.json"],
  "panel_dirs": ["config/panels"],
  "trusted_extension_dirs": []
}
```

## AI 工作流

AI 管理 panel 的理想流程：

1. 读取 `docs/generated/panel-skills.md` 和 workspace `config/daymine.json`。
2. 根据用户需求选择官方 renderer 或 preset。
3. 写入或修改 `config/panels/<id>.json`。
4. 更新 `config/pages/<page>.json` 的 columns。
5. 写入或更新对应数据文件，例如 `index/panels.json`、`notes/**/*.md`。
6. 调用验证 API 或命令，确认 manifest 引用、schema 和数据源都有效。

## 后端 API 建议

- `GET /api/v1/registry/renderers`：列出官方 renderer、字段契约和 presets。
- `GET /api/v1/dashboard/config`：返回已解析、已验证的页面树。
- `GET /api/v1/panels/:id`：返回 panel manifest + resolved data + renderer config。
- `POST /api/v1/panels/validate`：验证单个 panel DSL。
- `POST /api/v1/dashboard/reload`：让前端重新读取 workspace 配置。

## 实施顺序

1. 把现有硬编码 `type` 改成 `renderer.type` + `data.type`。
2. 后端实现 manifest loader 和 validation，保留向后兼容的 `config/daymine.json`。
3. 前端实现 renderer registry：`list`、`markdown`、`calendar`、`timeline`。
4. 实现 `html-template` renderer：解析 HTML fragment、白名单清洗、数据绑定、`dm-*` 标签映射。已完成首个 MVP 版本。
5. 把当前 `feed`、`article-list`、`github-list` 映射成 `list` preset。
6. 更新 `docs/generated/panel-skills.md`，让 AI 按新 DSL 生成 panel。
7. 增加 manifest 单测、template fixture、安全清洗测试和浏览器 smoke。

## 暂不做

- 默认运行第三方 JS。
- 远程下载插件。
- 让 panel 直接访问本地文件系统。
- 为每个外部平台都写专用 React 组件。
