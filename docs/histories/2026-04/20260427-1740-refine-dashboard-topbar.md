## [2026-04-27 17:40] | Task: Refine dashboard topbar

### 🤖 Execution Context

- **Agent ID**: `codex`
- **Base Model**: `gpt-5`
- **Runtime**: `Codex CLI`

### 📥 User Query

> 参考 Tididi，把 Daymine 页面里的顶 bar 改得更好看，去掉无用信息。

### 🛠 Changes Overview

**Scope:** `apps/web`, `docs`

**Key Actions:**

- **[Topbar layout]**: 将顶栏收敛为 Tididi 式品牌、页面导航和必要操作结构。
- **[Information cleanup]**: 移除顶栏里的本地 workspace 副标题和 `No auth` 状态展示。
- **[Navigation]**: 补上基于 dashboard pages 的页面切换能力，并在右侧保留当前页刷新入口。

### 🧠 Design Intent (Why)

Daymine 的 dashboard 应该优先服务反复扫描和操作。顶栏只保留可识别品牌、可切换页面和必要动作，避免把无操作价值的环境说明放在高频区域。

### 📁 Files Modified

- `apps/web/src/App.tsx`
- `apps/web/src/styles.css`
- `docs/FRONTEND.md`
