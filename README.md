# daymine

English version: [`daymine`](https://github.com/iFurySt/daymine)

## 简介

Daymine 是一个本地 self-hosted 的信息聚合、整理和长期回溯工具。它面向 Agent-first 的个人工作流：人给方向和日常检查，AI Agent 在本地 workspace 中整理资料、生成脚本、维护 Markdown/JSON 索引。

当前版本提供：

- Go 后端与内嵌 React 前端，一个二进制即可 serve。
- 无登录、无 OAuth、无多用户鉴权的本地单用户模式。
- 文件系统优先的 workspace，默认写入 Markdown、JSON 和 JSONL。
- 内置 dashboard panel：日历、feed、文章、GitHub、Agent runs、Markdown view。
- Agent Provider 抽象，首批支持 `local-command` 和 `codex-cli`。

## 快速开始

安装前端依赖并启动：

```bash
npm --prefix apps/web install
make run
```

默认访问：

- Web: `http://localhost:6345`
- API health: `http://localhost:6345/api/v1/health`
- Workspace: `~/.daymine/`，Windows 下对应当前用户 home 目录里的 `.daymine`

常用命令：

```bash
make test       # Go tests + frontend production build
make build      # Build embedded web assets and bin/daymine
make run        # Build web assets, then run the Go server
make ci         # Repository checks, Go tests, npm ci, frontend build
```

指定运行参数：

```bash
make run ADDR=:7345 WORKSPACE=/path/to/daymine-workspace
```

也可以直接运行：

```bash
go run ./apps/daymine/cmd/daymine --addr :6345
```

## 许可证

[MIT](LICENSE)

## 备注

架构和执行计划见 `docs/`，其中 `docs/exec-plans/active/2026-04-21-self-hosted-agent-dashboard.md` 跟踪当前开源 self-hosted 版本的推进状态。
