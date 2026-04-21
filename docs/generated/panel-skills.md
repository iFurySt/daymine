# Daymine Panel Skills

这份文档描述当前内置 panel contract，供 Agent 在 workspace 中生成或维护数据时参考。

## 数据入口

- Dashboard manifest: `config/daymine.json`
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

## Agent 写入规则

- 优先写 Markdown 长期内容，再更新 `index/panels.json` 的轻量索引。
- 新产物放入 `artifacts/runs/`、`artifacts/scripts/` 或 `artifacts/attachments/`。
- 不要写入登录、用户、token、OAuth 等本地版不需要的结构。
- 修改 manifest 后保持 `layout_by_column` 中引用的 panel id 都存在。
