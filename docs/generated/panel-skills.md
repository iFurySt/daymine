# Daymine Panel Skills

这份文档描述当前内置 panel contract，供 Agent 在 workspace 中生成或维护数据时参考。

长期目标见 `docs/design-docs/panel-plugin-system.md`：AI 应通过 Page DSL、Data Source DSL、HTML Template DSL 和 Renderer DSL 管理页面与 panel。当前 MVP 兼容简单的 `config/daymine.json` 单文件 manifest，并已支持 `html-template` 外置 panel。

## 数据入口

- Dashboard manifest: `config/daymine.json`
- HTML fragment templates: `config/panels/*.template.html`
- Panel-local CSS fragments: `config/panels/*.panel.css`
- Panel data index: `index/panels.json`
- Agent run log: `index/runs.jsonl`
- Markdown content: `notes/**/*.md`

## Manifest 结构

```json
{
  "pages": [
    {
      "name": "home",
      "title": "Home",
      "column_widths": [1, 1, 1],
      "layout_by_column": [["calendar", "feed"], ["article-list"], ["agent-runs"]],
      "panels": [
        {
          "id": "feed",
          "type": "feed",
          "title": "RSS Feed",
          "source": "index/panels.json"
        }
      ]
    }
  ]
}
```

默认 workspace 会包含一个外置 HTML template panel：`external-signal`。它由 `config/panels/external-signal.template.html`、`config/panels/external-signal.panel.css` 和 `index/panels.json.external-signal` 共同驱动。已有 workspace 如果缺少它，服务启动时会追加这个默认 panel 和示例数据。

## 内置 Panel

| Type | Data shape | Notes |
| --- | --- | --- |
| `calendar` | 后端生成 `{ today, events[] }` | 适合日程、每日复盘入口。 |
| `feed` | `index/panels.json.feed[]` | 每项建议包含 `title`, `source`, `summary`, `url`, `published_at`。 |
| `article-list` | `index/panels.json.article-list[]` | 每项建议包含 `title`, `path`, `status`, `tags`。 |
| `video-card` | `index/panels.json.<panel-id>[]` | 每项建议包含 `title`, `channel`, `url`, `summary`, `thumbnail`。 |
| `github-list` | `index/panels.json.github-list[]` | 每项建议包含 `name`, `full_name`, `description`, `stars`, `language`。 |
| `social-post` | `index/panels.json.<panel-id>[]` | 每项建议包含 `author`, `handle`, `text`, `url`, `published_at`。 |
| `agent-runs` | `index/runs.jsonl` | 由后端追加，不建议手写，除非在修复索引。 |
| `markdown-view` | `source` 指向 Markdown 文件 | 适合展示日报、主题页或 Agent 产物。 |
| `html-template` | `renderer.template_path` + `data` context | 适合 AI/用户在运行时持久化自定义 UI fragment。 |

## Agent 写入规则

- 优先写 Markdown 长期内容，再更新 `index/panels.json` 的轻量索引。
- 新产物放入 `artifacts/runs/`、`artifacts/scripts/` 或 `artifacts/attachments/`。
- 不要写入登录、用户、token、OAuth 等本地版不需要的结构。
- 修改 manifest 后保持 `layout_by_column` 中引用的 panel id 都存在。

## 目标 DSL 示例

未来官方 preset panel 应优先写成：

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
    "limit": 20
  }
}
```

未来动态自定义 panel 应优先使用 HTML fragment template：

```json
{
  "schema_version": "daymine.panel.v1",
  "id": "rss-html",
  "title": "RSS Feed",
  "renderer": {
    "type": "html-template",
    "template_path": "config/panels/rss.template.html"
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

Template fragment:

```html
<dm-list>
  <dm-item data-for="item in items">
    <dm-link href="{{ item.url }}">{{ item.title }}</dm-link>
    <dm-text tone="muted" max-lines="3">{{ item.summary }}</dm-text>
    <dm-meta>{{ item.source }} · {{ item.published_at }}</dm-meta>
  </dm-item>
</dm-list>
```

Template 规则：

- 写 panel body fragment，不写完整 `<html>` 页面。
- 用 `{{ path.to.value }}` 绑定后端 resolved context。
- 用 `data-for="item in items"` 循环。
- 用 `data-if="item.url"` 条件显示。
- 优先使用 `dm-*` 官方标签。
- 不写 `<script>`、`onclick`、`javascript:` URL 或外链脚本。

当前 MVP 支持的 `type`：`calendar`、`feed`、`article-list`、`github-list`、`agent-runs`、`markdown-view`、`html-template`。
