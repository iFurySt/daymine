## [2026-04-22 22:40] | Task: Panel plugin system design

### Execution Context

- **Agent ID**: `codex`
- **Base Model**: `GPT-5`
- **Runtime**: `Codex CLI`

### User Query

> 评估当前前端 panel 是否解耦，并设计一个最终单二进制运行后可由 AI 插拔管理页面和 panel 的机制。官方 panel 要内置，用户也能按官方规范自定义。

### Changes Overview

**Scope:** docs, packages, apps/web

**Key Actions:**

- **Design doc**: 新增 panel plugin system 设计，拆分 Page DSL、Data Source DSL、Renderer DSL 和官方 renderer registry。
- **Frontend docs**: 更新 panel 边界，明确 `feed`/`github-list` 等应收敛为 `list` preset，而不是长期硬编码 renderer。
- **Agent docs**: 更新 panel skills，让 AI 知道当前 MVP contract 和目标 DSL。
- **Plan update**: 将 renderer registry 和 manifest validation 纳入 active execution plan。
- **HTML template direction**: 补充类似 Go `html/template` 的 HTML fragment panel 设计，明确数据 context、`{{ }}` 绑定、`data-for`/`data-if`、`dm-*` 官方标签和受控 CSS 边界。
- **MVP implementation**: 增加默认 `external-signal` 外置 panel，后端解析 `renderer`/`data`，前端运行时渲染 HTML fragment。
- **Workspace migration**: 启动时为缺少 `external-signal` 的既有 workspace 追加默认 panel 和示例数据，不覆盖已有 panel。

### Design Intent (Why)

当前 MVP 已经让 AI 能改 workspace manifest 和数据，但 renderer 仍写死在前后端代码里。设计目标是让二进制内置官方 renderer，同时允许 AI 通过受控 DSL 管理页面、数据源、字段映射和动作；自定义 panel 先走 HTML fragment template、schema/preset，不默认执行第三方 JS。

实现阶段先保留内置 panel 的兼容路径，同时新增 `html-template` renderer 跑通动态自定义 UI。这样用户可以继续使用内置 panel，也能通过 workspace template/style/data source 验证外置 panel 的实际效果。

### Files Modified

- `docs/design-docs/panel-plugin-system.md`
- `docs/design-docs/index.md`
- `docs/ARCHITECTURE.md`
- `docs/FRONTEND.md`
- `docs/generated/panel-skills.md`
- `docs/exec-plans/active/2026-04-21-self-hosted-agent-dashboard.md`
- `packages/workspace/workspace.go`
- `packages/panels/service.go`
- `packages/panels/service_test.go`
- `apps/web/src/App.tsx`
- `apps/web/src/types.ts`
