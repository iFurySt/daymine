# Self-Hosted Agent Dashboard 开源版执行计划

## 目标

把 daymine 从 Agent-first 仓库模板推进成一个本地 self-hosted 信息聚合与长期回溯工具：Go 后端提供单二进制服务和 Agent 控制层，React 前端复刻 tididi 的紧凑 dashboard 风格，数据默认落本地文件系统，移除登录鉴权体系，并为内置/扩展 panel 建立稳定契约。

## 范围

- 包含：
  - 从 `/Users/ifuryst/projects/amoylab/tididi` 迁移可复用的视觉风格、dashboard 布局思想和 widget 形态。
  - 设计并实现无登录的本地单用户运行模式。
  - 建立 Go 后端、内嵌 React 前端、workspace FS 存储、Agent Provider 抽象和 panel registry。
  - 为每个代码切片补测试，并在文档和 history 中记录完成状态。
- 不包含：
  - 多租户 SaaS、注册登录、OAuth、邮箱验证、管理员权限。
  - 强依赖 Postgres 或远程数据库的部署方式。
  - 一开始就支持所有 Agent 厂商的完整能力；先用统一接口和 1-2 个 provider 验证边界。

## 背景

- 相关文档：
  - `docs/product-specs/self-hosted-agent-dashboard.md`
  - `docs/ARCHITECTURE.md`
  - `docs/FRONTEND.md`
  - `docs/PRODUCT_SENSE.md`
- 相关代码路径：
  - 目标仓库：`apps/`, `packages/`, `docs/`
  - 参考项目：`/Users/ifuryst/projects/amoylab/tididi/cmd/server`, `internal/server`, `internal/service`, `web/src`
- 已知约束：
  - 生产交付优先单 Go 二进制，内嵌前端静态资源。
  - 数据优先文件系统和 Markdown，SQLite 只能作为可选控制/索引层。
  - Agent 后端不能绑定到单一实现，核心只管理输入、执行、状态、日志和产物。
  - UI 风格应继承 tididi 的紧凑 dashboard，而不是重做营销页或重型后台。

## 风险

- 风险：直接复制 tididi 会把 Auth/Postgres/GORM 用户配置耦合带进来。
  - 缓解方式：先建立 daymine 自己的 workspace、panel 和 agent contract，再按文件迁移可复用代码。
- 风险：纯文件系统查询会在数据量增长后变慢。
  - 缓解方式：把 append-only 事件和 JSON 索引作为第一阶段能力，SQLite/FTS 作为可替换索引实现。
- 风险：Agent Provider 抽象过早泛化导致无法落地。
  - 缓解方式：第一阶段只实现 `local-command` 和 `codex-cli`，用真实 run 记录反推接口。
- 风险：React 插件化如果直接允许任意 HTML/JS，会带来安全和维护成本。
  - 缓解方式：先采用内置 React panel + manifest/data contract；扩展从模板和技能文档开始。

## 里程碑

1. 方案收敛与仓库落点。
   - 更新产品规格、架构、前端规范和质量评分。
   - 明确从 tididi 继承与替换的边界。
2. 项目骨架。
   - 新增 Go module、HTTP server、配置加载、workspace 初始化和健康检查。
   - 新增 React/Vite 前端，迁入 tididi 视觉变量和基础布局。
   - 配置 Go embed，把前端 build 产物打进二进制。
3. 本地数据层。
   - 实现 workspace 目录结构、Markdown/JSON/YAML 读写、run JSONL 事件记录。
   - 实现 panel manifest loader、默认示例数据和 schema 校验。
   - 评估是否需要 SQLite 作为可选索引层。
4. Panel 系统。
   - 实现 panel registry 和 API：calendar、feed、article-list、video-card、github-list、social-post、agent-runs、markdown-view。
   - 生成面向 Agent 的 panel skills/template 文档。
   - 前端支持多页、多列、刷新、错误态和空态。
5. Agent 控制层。
   - 定义 provider 接口、run lifecycle、日志和 artifact discovery。
   - 实现 `local-command` provider 和 `codex-cli` provider。
   - 前端提供 query 输入、运行状态、产物链接和重试。
6. 开源发布准备。
   - README、配置示例、二进制构建、CI、release package、供应链检查。
   - 基础 e2e smoke、跨平台构建和文档同步。

## 验证方式

- 命令：
  - `make check-docs`
  - `make check-repo`
  - `go test ./...`
  - `npm run build` 或仓库封装后的前端构建命令
  - `go test ./... && go build ./cmd/daymine`
- 手工检查：
  - 本地启动二进制后无需登录进入 dashboard。
  - workspace 初始化内容可用编辑器/Obsidian 直接阅读。
  - Agent run 后能看到状态、日志摘要和产物文件。
- 观测检查：
  - 每次 Agent run 都有 run id、开始/结束时间、provider、输入摘要、输出文件和失败原因。
  - 后端日志能定位 panel load、workspace IO 和 provider 执行错误。

## 进度记录

- [x] 调研 tididi 顶层结构、Go 服务入口、Auth 耦合、dashboard 前端和 widget 注册方式。
- [x] 建立 self-hosted 开源版产品规格。
- [x] 更新 daymine 架构、前端规范、产品判断和质量评分。
- [x] 初始化 Go + React 项目骨架。
- [x] 移除登录路径并实现无鉴权 dashboard shell。
- [x] 实现 workspace FS 数据层和默认示例数据。
- [x] 实现首批 panel registry 与 API。
- [x] 实现 Agent Provider 接口和首个 provider。
- [x] 补齐 Go 单测、前端构建和仓库 CI 串联。
- [x] 把 release package 从模板元数据包替换成真实二进制打包。
- [x] 设计 panel plugin/DSL：Page DSL、Data Source DSL、HTML Template DSL、Renderer DSL、官方 renderer registry 和自定义 panel 层级。
- [x] 跑通首个外置 `html-template` panel：workspace template/style/data source，前端运行时绑定渲染。
- [ ] 把当前硬编码 panel renderer 迁移到 renderer registry 和 manifest validation。
- [ ] 补浏览器 smoke、真实信息源导入和更完整的 provider 配置。

## 决策记录

- 2026-04-21：生产交付形态定为单 Go 二进制内嵌 React 前端，降低 self-hosted 部署门槛。
- 2026-04-21：核心数据采用文件系统优先，Markdown 面向人类长期阅读，JSON/YAML/JSONL 面向机器索引和 Agent 操作记录。
- 2026-04-21：不迁移 tididi 的鉴权、OAuth、邮件验证和 Postgres 用户配置链路，避免开源本地版背上 SaaS 架构成本。
- 2026-04-21：Panel 扩展先采用 manifest + 内置 React registry，而不是直接运行第三方 HTML/JS；后续再根据真实需求开放更强插件能力。
- 2026-04-21：MVP 采用标准库 HTTP server 和手写 CSS，先减少依赖面；前端生产构建写入 Go embed 目录，由 `make build` 和 `make run` 统一生成。
- 2026-04-21：`local-command` provider 先作为本地可信执行入口保留，但安全文档明确它不适合公网暴露和非可信输入。
- 2026-04-21：默认 workspace 改为用户 home 下的 `.daymine`，Windows/macOS/Linux 都通过 Go 的 `os.UserHomeDir()` 解析；仍保留 `--workspace` 和 `make run WORKSPACE=...` 覆盖。
- 2026-04-22：Panel 目标架构采用受控 DSL，而不是默认执行第三方 JS。官方 renderer 内置在二进制中，社区自定义优先通过 preset、schema renderer 和 workspace manifest 实现。
- 2026-04-22：动态自定义 panel 采用类似 Go `html/template` 的 HTML fragment template：后端 resolve data context，前端做受控绑定和 `dm-*` 官方标签映射；允许 GENUI 风格块级 HTML，但禁止任意脚本。
- 2026-04-22：默认 workspace 增加 `external-signal` 外置 panel，用 `config/panels/external-signal.template.html` 和 `index/panels.json` 验证内置 panel 与外置 HTML template panel 可共存。
