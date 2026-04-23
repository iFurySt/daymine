# 质量评分

用这份文档按产品区域和架构层次记录当前质量水位，方便持续知道最薄弱的地方在哪。

## 建议的评分标准

- `A`：覆盖完整、行为稳定、文档清楚、运行风险低。
- `B`：整体可接受，但还有明确短板。
- `C`：能用，但需要针对性补强。
- `D`：脆弱、缺少规范，或很多行为尚未定义。

## 当前评分

| 区域 | 评分 | 原因 | 下一步 |
| --- | --- | --- | --- |
| 产品面 | B | 已有可运行 MVP：无登录 dashboard、默认 workspace、内置 panel、Agent run 入口和首个 HN task case。 | 补更多真实信息源导入、搜索和长期回溯视图。 |
| 架构文档 | B | 已替换成 Go + React + FS + Agent Provider + panel registry 的目标架构，并有对应代码骨架。 | 随接口稳定补 API contract 和 package 依赖图。 |
| 前端 | B | React dashboard 已接入 API，复刻 tididi 的紧凑深色风格，支持默认 panel 和 Agent run。 | 增加浏览器 smoke、更多 panel 交互和移动端细节验证。 |
| 数据层 | B | 已实现 workspace layout、默认 Markdown/JSON、panel index 和 run JSONL。 | 增加 schema 校验、备份/恢复和可选 SQLite/FTS 评估。 |
| Agent 控制层 | C | 已有 provider 接口、`local-command`、`codex-cli`、task registry、每日调度开关和 artifact discovery。 | 加强命令执行边界、取消/重试、provider 配置和 scheduler 可观测性。 |
| 测试 | C | Go 单测覆盖核心包，前端构建纳入 `make ci`。 | 增加 Playwright smoke 和 API contract 测试。 |
| 可观测性 | B | HTTP 结构化日志和 Agent run history 已落地。 | 增加前端可见日志详情和 panel load 错误分类。 |
| 安全 | C | 明确本地无鉴权和 `local-command` 风险，默认 workspace 边界已形成。 | 增加 provider allowlist、环境变量白名单和禁用高风险 provider 的配置。 |
