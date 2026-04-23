# 安全默认约束

Daymine 当前定位是本地 self-hosted 单用户工具，默认移除登录、注册、OAuth、邮箱验证和多租户权限体系。这个选择降低部署复杂度，但也意味着服务不应该直接暴露到公网。

## 默认边界

- 默认监听地址由启动参数控制，开发建议绑定 `127.0.0.1:6345` 或受信任内网。
- 无鉴权 API 只适合本机或可信网络使用。
- Workspace 是核心数据边界，默认位于用户 home 目录下的 `.daymine`；服务只应读写启动时指定的 workspace。
- Markdown、JSON、JSONL 和 artifact 都可能包含用户资料，不应写入 history 或公开日志。

## Agent 执行

- `local-command` provider 会在 workspace 目录中执行用户提交的 shell 命令，只适合可信本机使用。
- `codex-cli` provider 调用本机 `codex exec --full-auto --skip-git-repo-check --cd <workspace>`，认证和密钥由用户本机 Codex CLI 管理；任务 prompt 必须把允许写入的 workspace 路径写清楚。
- Agent run 必须记录 provider、query、状态、开始/结束时间、输出摘要、失败原因和产物路径。
- `--scheduler` 会自动触发到期 task，默认 `make run` 不启用；启用前应确认对应 task/provider 的写入范围和网络访问风险。
- 后续加强方向：命令 allowlist、环境变量白名单、工作目录限制、超时配置、可选禁用 `local-command`。

## 依赖与供应链

仓库级的依赖、SBOM 和 provenance 默认能力，统一写在 `docs/SUPPLY_CHAIN_SECURITY.md`。
