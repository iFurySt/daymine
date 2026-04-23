## [2026-04-22 17:30] | Task: Codex task Hacker News

### 🤖 Execution Context

- **Agent ID**: `codex`
- **Base Model**: `gpt-5`
- **Runtime**: `Codex CLI`

### 📥 User Query

> 抽象一个能直接运行 Codex 做任务的能力，参考 Happy 的 `happy codex` wrapper；先用最近一天 Hacker News Top 10 作为 case，后续可每天调度并累积知识，在前端可见。

### 🛠 Changes Overview

**Scope:** `packages/agent`, `packages/tasks`, `packages/workspace`, `packages/panels`, `apps/daymine`, `apps/web`, `docs`

**Key Actions:**

- **[Task abstraction]**: 新增 task registry/service，把 task id、provider、prompt、schedule、artifact 和 panel 关联收口。
- **[Codex provider]**: 将 `codex-cli` provider 改为 `codex exec --full-auto --skip-git-repo-check --cd <workspace> -`，用 stdin 传入任务 prompt。
- **[HN case]**: 新增 `hacker-news-daily-top10` task 和 `hacker-news-top` panel，产物落到 `index/hacker-news/` 与 `notes/sources/hacker-news/`。
- **[UI/API]**: 增加 task list/run API，前端 HN panel 可触发 task 并展示最新 digest。
- **[Scheduler]**: 增加 `--scheduler` 启动开关，启用后按每日语义检查并执行到期 task。

### 🧠 Design Intent (Why)

Happy 的完整 wrapper 管理长期 session、远程控制、权限和事件流；Daymine 现阶段先落更小的 provider-agnostic task 层，让主 Agent 只需要管理「输入、执行、产物、展示」这条稳定链路，Codex 只是其中一个 provider。

### 📁 Files Modified

- `packages/tasks/service.go`
- `packages/agent/agent.go`
- `packages/workspace/workspace.go`
- `packages/panels/service.go`
- `apps/daymine/internal/server/server.go`
- `apps/daymine/cmd/daymine/main.go`
- `apps/web/src/App.tsx`
- `apps/web/src/api.ts`
- `apps/web/src/types.ts`
- `apps/web/src/styles.css`
- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/FRONTEND.md`
- `docs/SECURITY.md`
- `docs/RELIABILITY.md`
- `docs/product-specs/self-hosted-agent-dashboard.md`
