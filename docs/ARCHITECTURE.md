# 架构总览

Daymine 是一个本地 self-hosted 的信息聚合、整理和长期回溯工具。核心模式是：用户给方向和日常检查，AI Agent 负责收集、整理、生成脚本和维护 workspace；Go 服务负责控制执行、读写本地文件系统、提供 API，并用内嵌 React 前端展示 dashboard。

## 运行拓扑

```text
Browser
  |
  v
Go HTTP server
  |-- embedded React assets
  |-- REST/Event API
  |-- workspace filesystem store
  |-- panel registry
  |-- agent controller
        |-- local command provider
        |-- codex cli provider
        |-- future coding-agent providers
```

生产形态优先是一个二进制：启动后 serve API 和前端静态资源。开发形态可以拆成 Go server + Vite dev server。

## 预期仓库结构

- `apps/daymine/`：Go 服务入口、HTTP API、前端内嵌和本地运行命令。
- `apps/web/`：React + TypeScript 前端，开发期由 Vite 管理，生产期 build 后被 Go embed。
- `packages/agent/`：Agent Provider 接口、run lifecycle、日志和产物发现。
- `packages/workspace/`：本地 workspace 初始化、文件读写、索引和 schema 校验。
- `packages/panels/`：panel manifest、registry、默认 panel contract 和示例数据。
- `infra/`：后续的 Docker、release、打包和 provenance。
- `scripts/`：仓库级自动化脚本，供人和 Agent 直接调用。
- `docs/`：仓库知识库，也是本地规则和上下文的正式来源。

实际落地时可以按 Go module 的惯例调整目录名，但依赖边界要保持清楚。

## 数据流

1. 用户在前端查看 panel 或提交 query。
2. Go API 读取 workspace 的配置、索引和 Markdown/JSON 数据。
3. Agent controller 创建 run id，调用配置好的 provider。
4. Provider 在受控 workspace 中执行，产出消息、日志和文件。
5. Controller 记录 append-only run 事件，并刷新 panel 索引。
6. 前端通过 API 重新读取 panel 数据和 run 状态。

## 存储模型

- 文件系统是事实来源，默认 workspace 结构见 `docs/product-specs/self-hosted-agent-dashboard.md`。
- Markdown 存长期内容，JSON/YAML 存配置和索引，JSONL 存 Agent run 事件。
- SQLite 只作为可选控制/索引层，不应成为核心内容唯一来源。

## 关键边界

- 无鉴权：本地版默认不提供登录、注册、OAuth、邮箱验证和多用户权限。
- Agent 可替换：核心只管理输入、执行状态、日志和产物，不把业务逻辑绑死到 Codex 或某个 LLM API。
- Panel 可组合：前端使用内置 React renderer registry、manifest contract 和受控 DSL，不让任意 HTML/JS 直接获得无边界执行权。
- 文档同步：架构、panel contract、provider contract 或 workspace layout 变化时，同步更新 `docs/` 和对应 history。

Panel 插件系统的目标设计见 `docs/design-docs/panel-plugin-system.md`。当前 MVP 仍有部分 renderer 和 data source 硬编码，后续要收敛到 Page DSL、Data Source DSL、Renderer DSL 和内置 renderer registry。
