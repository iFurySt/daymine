## [2026-04-22 11:50] | Task: 修复 GitHub Actions Markdown lint

### Execution Context

- **Agent ID**: `Codex`
- **Base Model**: `GPT-5`
- **Runtime**: `Codex CLI`

### User Query

> gh 到 GitHub Action 看看为什么之前失败了，并修复。

### Changes Overview

**Scope:** GitHub Actions CI

**Key Actions:**

- **[Diagnose]**: 使用 `gh run view` 查看失败 CI，确认 `markdownlint-cli2` 扫描了 `apps/web/node_modules/**/*.md`。
- **[Fix]**: 在 CI 的 Markdown lint glob 中排除所有 `node_modules` 目录，避免第三方依赖文档影响仓库门禁。

### Design Intent (Why)

CI 需要检查仓库维护的 Markdown 文档，不应该在 `npm ci` 后把下载到本地的依赖包文档纳入 lint 范围。排除 `node_modules` 可以保留现有 Markdown 规则，同时让 workflow 对前端依赖安装顺序保持稳定。

### Files Modified

- `.github/workflows/ci.yml`
- `docs/histories/2026-04/20260422-1150-fix-markdownlint-node-modules.md`
