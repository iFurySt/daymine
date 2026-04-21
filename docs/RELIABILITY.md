# 稳定性与可运维性

Daymine 的早期可靠性目标是本地可启动、可诊断、可重复验证。每个 Agent run 和 panel load 都应该能从本地文件和日志中追踪。

## 当前底线

- `GET /api/v1/health` 返回服务健康状态。
- 启动时自动初始化 workspace 目录和默认示例数据。
- HTTP 请求使用结构化日志记录 method、path 和 duration。
- Agent run 默认有超时，结果追加到 `index/runs.jsonl`。
- 前端生产构建产物内嵌到 Go 二进制，刷新路由 fallback 到 `index.html`。

## 验证路径

- `go test ./...` 覆盖 workspace、panel、agent 和 server 基础行为。
- `npm --prefix apps/web run build` 覆盖 React/TypeScript 构建和 embed 产物生成。
- `make ci` 串起仓库卫生、shell 语法、Go 测试、`npm ci` 和前端构建。

## 后续加强

CI/CD 流程结构和 release 自动化的默认方案，统一写在 `docs/CICD.md`。

- 增加浏览器 smoke 测试，验证 dashboard 无登录可见。
- 增加 Agent run 取消、重试和更细的错误分类。
- 增加 panel 数据 schema 校验和损坏索引恢复提示。
