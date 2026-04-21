## [2026-04-21 23:34] | Task: Self-hosted agent dashboard planning

### Execution Context

- **Agent ID**: `codex`
- **Base Model**: `GPT-5`
- **Runtime**: `Codex CLI`

### User Query

> 调研本地 tididi 项目，规划开源 self-hosted 版本：复刻前端风格，移除鉴权登录，采用 Go + React 单二进制交付，文件系统优先存储，抽象 Codex/Claude Code 等 Agent Provider，并把 panel 组件化和长期 TODO 落到仓库。

### Changes Overview

**Scope:** docs, apps, packages, scripts

**Key Actions:**

- **Product spec**: 新增 self-hosted agent dashboard 产品规格，明确目标用户、数据模型、Agent 抽象、panel 系统和 tididi 继承边界。
- **Execution plan**: 新增 active plan，拆解 Go/React 骨架、FS 数据层、panel registry、Agent 控制层和开源发布准备。
- **Architecture docs**: 更新架构、前端协作、产品判断和质量评分，从模板状态切换到 daymine 的真实方向。
- **Go MVP**: 新增 workspace、panel、agent provider、HTTP server 和内嵌 web assets。
- **React MVP**: 新增无登录 dashboard，展示默认 panel，并提供本地 Agent run 表单。
- **Verification**: 接入 `make test`, `make build`, `make run`, `make ci`，并把 Go 单测和前端构建纳入 CI 脚本。
- **Release package**: 将 release 脚本从模板元数据包改成构建前端、构建 Go 二进制并打包本机平台 artifact。
- **Workspace default**: 将默认 workspace 从当前目录 `.daymine` 改为用户 home 目录下的 `.daymine`，并保留显式覆盖参数。

### Design Intent (Why)

先把长期方向和边界版本化，避免后续直接复制 tididi 时把旧的 Auth/Postgres/SaaS 假设带入本地开源版。计划把可复用的视觉和 dashboard 经验留下，把登录、用户表、OAuth 和远程数据库依赖替换成 workspace、文件系统、panel manifest 和 Agent Provider contract。

实现阶段继续保持这个边界：后端不用数据库和鉴权，workspace 是事实来源；前端直接进入 dashboard；Agent 控制层只围绕 query、provider、运行记录和产物路径建模。

### Files Modified

- `docs/product-specs/self-hosted-agent-dashboard.md`
- `docs/product-specs/index.md`
- `docs/exec-plans/active/2026-04-21-self-hosted-agent-dashboard.md`
- `docs/ARCHITECTURE.md`
- `docs/FRONTEND.md`
- `docs/PRODUCT_SENSE.md`
- `docs/QUALITY_SCORE.md`
- `README.md`
- `Makefile`
- `scripts/ci.sh`
- `scripts/release-package.sh`
- `go.mod`
- `apps/daymine/cmd/daymine/main.go`
- `apps/daymine/internal/server/server.go`
- `apps/daymine/internal/webassets/assets.go`
- `apps/web/`
- `packages/agent/`
- `packages/panels/`
- `packages/workspace/`
- `docs/generated/panel-skills.md`
- `docs/SECURITY.md`
- `docs/RELIABILITY.md`
