# Self-Hosted Agent Dashboard 产品规格

## 定位

Daymine 是一个本地 self-hosted 的长期信息聚合、整理和回溯工具。它面向个人用户和小型团队：人负责给方向、每天浏览和校正，AI Agent 负责持续收集、整理、生成脚本、维护资料结构。

## 核心价值

- 长期可回溯：信息不只展示当天结果，也要能按日期、主题、来源和 Agent 操作记录回看。
- 本地可拥有：默认单机运行，数据落在本地文件系统，优先使用 Markdown、JSON/YAML 和附件目录，方便 Obsidian、编辑器和备份工具读取。
- Agent-first：系统把任务输入、执行上下文、产物路径和执行结果管理好；具体执行可以由 Codex、Claude Code、OpenClaw、Hermes、LLM API 或其他 Coding Agent 完成。
- 开箱即用：生产形态是一个 Go 二进制，内嵌打包好的 React 前端，不要求用户部署数据库或额外 Web 服务器。

## 明确不做

- 不提供多用户登录、注册、OAuth、邮箱验证、权限管理和 SaaS 租户体系。
- 不把数据强绑定到远程数据库。
- 不把某一个 Agent 厂商写死到核心业务模型里。
- 不优先做复杂在线协作；本地文件可同步，但冲突处理先交给用户选用的同步方案。

## 目标用户路径

1. 用户下载或构建一个二进制。
2. 用户启动服务；默认 workspace 位于用户 home 目录下的 `.daymine`，也可以显式指定其他目录。
3. 前端展示默认 dashboard：日历、信息流、文章、视频、GitHub、RSS、推文/社媒摘录、Agent 任务与最近产物。
4. 用户输入 query 或选择模板任务，系统调用已配置的 Agent Provider。
5. Agent 读取 workspace、生成或更新 Markdown/JSON/脚本，系统记录执行事件和产物索引。
6. 用户每天查看聚合面板，并可按日期、来源、主题和 Agent run 回溯。

## 数据模型

### 文件系统优先

默认 workspace 结构建议：

```text
workspace/
  config/
    daymine.yaml
    panels/
  inbox/
    rss/
    web/
    social/
    manual/
  notes/
    daily/
    topics/
    sources/
  artifacts/
    runs/
    scripts/
    attachments/
  index/
    panels.json
    sources.json
    runs.jsonl
```

规则：

- Markdown 用于人可读的长期内容。
- JSON/YAML 用于机器可读的索引、panel 配置和运行状态。
- 大文件、截图、视频封面、导出包放在 `artifacts/attachments/`。
- Agent run 记录采用 append-only JSONL，便于审计和恢复。

### SQLite 的边界

SQLite 可以作为可选控制层，用于缓存、全文索引、队列状态和快速查询；核心资料仍以文件为准。早期优先避免 cgo 绑定：如果需要 SQLite，先评估纯 Go driver、构建体积、跨平台二进制和 FTS 能力，再落实现。

## Agent Provider 抽象

核心接口围绕输入和产物，不围绕某个模型 API：

- 输入：query、workspace 路径、允许读写范围、模板、环境变量引用、超时和期望输出。
- 输出：结构化消息、退出码、日志摘要、生成/修改的文件路径、可展示 artifact、失败原因。
- Provider：Codex CLI、本地命令、Claude Code、OpenClaw/Hermes、直接 LLM API、未来远程 runner。
- 控制层：负责 run id、工作目录、日志、超时、中断、产物发现、状态事件和前端订阅。

## Panel 与组件系统

内置 panel 先覆盖常见信息形态：

- `calendar`：日期、日程、日总结入口。
- `feed`：RSS/Atom/网页抓取条目。
- `article-list`：长文、摘要、稍后读、主题列表。
- `video-card`：视频条目、频道、封面、摘要、观看状态。
- `github-list`：仓库、release、issue、star、趋势。
- `social-post`：X/Twitter 等短文本摘录，先以导入数据或 Agent 采集产物为准。
- `agent-runs`：最近任务、状态、产物和失败重试。
- `markdown-view`：直接渲染 workspace 中的 Markdown 片段。

组件化方向：

- React 内置组件负责交互密度、主题、筛选和状态。
- Panel manifest 用 JSON/YAML 描述数据源、字段、布局、刷新策略和可用动作。
- 对外暴露 `docs/generated/panel-skills.md` 或同类文件，告诉 Agent 当前有哪些 panel、字段契约和模板。
- 第三方扩展优先从静态 manifest + Markdown/JSON 数据开始；只有确实需要复杂交互时才接入前端插件 API。

## 从 tididi 继承

可以继承：

- 深色、紧凑、monospace、低圆角的 dashboard 视觉风格。
- React + TypeScript + Tailwind + shadcn/radix 风格基础组件。
- 多页面、多列、组件 id + type + config 的 dashboard 配置模型。
- Go HTTP 服务、配置加载、scheduler 和 scraper 的分层思路。

必须替换：

- AuthProvider、ProtectedRoute、JWT、OAuth、用户模型、邮件验证。
- Postgres/GORM 作为核心数据来源的假设。
- 用户维度 dashboard 配置，改为 workspace 维度配置。
- 固定 widget API，改为 panel registry + file-backed data source。

## 首批验收标准

- 一个 Go 二进制能 serve API 和内嵌前端。
- 无需登录即可进入 dashboard。
- 默认 workspace 初始化后包含示例 Markdown、panel manifest 和 run 记录。
- 前端能展示至少 `calendar`、`feed`、`article-list`、`github-list`、`agent-runs` 五类 panel。
- 可以通过一个 Agent Provider 适配器执行本地命令或 Codex CLI，并把输出文件纳入索引。
- 仓库有 Go 单元测试、前端构建检查和最小端到端 smoke 验证。
