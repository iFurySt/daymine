## [2026-04-27 15:23] | Task: Remove supply-chain workflow

### 🤖 Execution Context

- **Agent ID**: `codex`
- **Base Model**: `gpt-5`
- **Runtime**: `Codex CLI`

### 📥 User Query

> 提交相关改动，并移除当前每次推送都会报错且暂时没用的供应链安全 CI。

### 🛠 Changes Overview

**Scope:** `.github`, `scripts`, `docs`

**Key Actions:**

- **[Workflow removal]**: 删除常驻的 `supply-chain-security.yml`，停止 PR、定时和手动触发的 dependency review / OSV 扫描。
- **[Config cleanup]**: 删除已无调用方的 dependency review 配置，并从仓库卫生检查必需文件列表移除。
- **[Docs]**: 更新 CI/CD 和供应链安全文档，说明当前只保留 release SBOM/provenance 和 action pinning，PR/OSV 扫描后续按需恢复。

### 🧠 Design Intent (Why)

当前项目仍处在模板早期阶段，依赖和误报处理流程尚未稳定。常驻供应链扫描已经变成推送阻塞项，因此先移除常驻 workflow，保留 release 侧可追溯产物能力。

### 📁 Files Modified

- `.github/workflows/supply-chain-security.yml`
- `.github/dependency-review-config.yml`
- `scripts/check-repo-hygiene.sh`
- `docs/CICD.md`
- `docs/SUPPLY_CHAIN_SECURITY.md`
