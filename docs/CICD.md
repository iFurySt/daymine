# CI/CD 说明

这个模板自带一套不依赖具体语言栈的 CI/CD 骨架。

## 默认包含的内容

- `ci.yml`：仓库级检查，覆盖 docs、repo hygiene、Markdown 和 shell 脚本校验。
- `supply-chain-security.yml`：在 PR 上做依赖变更检查，并在 PR、定时任务和手动触发时运行 OSV 扫描。
- `release.yml`：手动触发的 release 流水线，用来打包仓库级制品、生成 provenance，并创建 GitHub Release。

## 设计原则

这套流水线的目标，是持续产出可追溯的本地 self-hosted 制品。

`scripts/release-package.sh` 当前会安装前端依赖、构建 React 静态资源、构建 Go 二进制，并打包成本机平台的 `daymine-${GOOS}-${GOARCH}.tgz`。

所有 GitHub Actions 都已经 pin 到 commit SHA。后续升级 action 时，也要继续保持这个约束。

## 推荐接入顺序

1. 保留 `ci.yml`，作为唯一默认常驻的仓库基础门禁。
2. 在 `scripts/ci.sh` 里继续叠加项目自己的验证命令。
3. 用真实构建产物替换 `scripts/release-package.sh`。
4. 技术栈和环境稳定后，再补具体的部署 job。
5. 即使交付方式变化，SBOM 和 provenance 这类供应链能力也建议保留。

## 默认 release 产物

当前 release 流水线会产出：

- `release-manifest.json`
- `daymine-${GOOS}-${GOARCH}.tgz`
- `sbom.spdx.json`
- 对 release artifact 生成的 GitHub artifact attestation

当前脚本先构建当前 runner 平台的二进制。后续如果要发布多平台制品，应在 release workflow 中矩阵化 `GOOS/GOARCH`，并为每个平台分别生成 artifact。
