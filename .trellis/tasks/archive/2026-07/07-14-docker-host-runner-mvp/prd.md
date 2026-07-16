# 实现 Docker Host Runner MVP

## Goal

为 Fork 增加面向 Mac mini + OrbStack 的 `docker-host` Backend：收到 GitHub Runner Scale Set 任务后按需启动单个 Linux AMD64 JIT Runner 容器，通过活动 Docker Context 的 Unix Socket 使用宿主 Docker Engine，任务完成后删除容器和匿名卷，同时保留明确配置的构建缓存。

## Background

- 基线提交为 `d0197af`，开发分支为 `feat/docker-host-runner`。
- 现有 `docker` Backend 保持 privileged DinD 语义，不做兼容性替换。
- 本机活动 Docker Context 为 `orbstack`，Endpoint 为动态解析得到的 Unix Socket；不得硬编码个人路径。
- 功能实现和单元测试期间不使用真实 GitHub 凭证；提交后使用仓库外的最小权限 Fine-grained PAT 执行一次真实 JIT Smoke。
- 上游基线 `make unit-test` 在现有 Tart 测试 `TestWaitForSSHReportsLastReadinessError` 失败，其他已执行包通过。

## Requirements

### R1 Backend and scope

- 新增独立 Backend 名称 `docker-host`，实现现有 `backend.Backend` 的 `Run`、`Cleanup`、`CleanupAll`。
- 第一阶段只允许 Repository-level `docker-host` Runner Set、`max_runners: 1`、`platform: linux/amd64`。
- 不改变或删除 `docker`、`tart` Backend。

### R2 Docker context and runner launch

- 优先支持显式 `DOCKER_HOST`；未设置时通过 `docker context show` 与 `docker context inspect` 解析活动 Context。
- 只接受绝对路径 `unix://` Endpoint；错误必须包含 Runner、镜像、Platform、Context 等诊断信息。
- `docker run` 使用 detached 模式，挂载 Socket 到 `/var/run/docker.sock`，不使用 `--privileged`。
- 通过进程环境转发 `ACTIONS_RUNNER_INPUT_JITCONFIG`，命令参数和错误日志不得包含完整 JIT Config。
- 容器带有 EFR 管理标签、Runner Set 标签和 Runner 名称标签。
- OrbStack Socket 在容器内表现为 `root:root 0660` 时，Runner 用户必须获得可访问该 Socket 的附加组。

### R3 Cleanup safety

- `Cleanup` 先使用管理标签、Runner Set 和 Runner 名称精确筛选，再执行等价于 `docker rm -fv` 的删除；不存在时成功返回。
- `CleanupAll` 只使用 EFR 管理标签与 Runner Set 标签筛选遗留容器，不依赖模糊名称匹配。
- Cleanup 不删除命名缓存卷、宿主缓存目录或 BuildKit Cache。

### R4 Runner image

- 新镜像位于 `images/docker-host-runner/`，基于固定版本 `ghcr.io/actions/actions-runner:2.332.0`。
- 镜像保留 Git、curl、jq、Python 3、CA 证书、Docker CLI 和 Buildx，增加固定版本 Compose CLI。
- 删除 Docker Daemon 相关可执行文件，不启动 Dockerd；入口执行 `/home/runner/run.sh` 并支持 JIT Config 环境变量。
- 镜像不包含仓库凭证，只要求 `linux/amd64`，为后续 ARM64 保留清晰扩展方式。

### R5 Cache layout

- 全局配置增加绝对路径 `cache_root`，`docker-host` Runner Set 增加安全的单段 `cache_namespace`。
- 宿主目录挂载布局：
  - `shared/cargo-home` -> `/home/runner/.cargo`
  - `<namespace>/cargo-target` -> `/home/runner/.cache/efr/cargo-target`
  - `<namespace>/sccache` -> `/home/runner/.cache/sccache`
  - `shared/pnpm-store` -> `/home/runner/.cache/pnpm-store`
  - `shared/tool-cache` -> `/opt/hostedtoolcache`
- Runner 工作目录、JIT Config 和 Token 不进入缓存目录。
- BuildKit 使用宿主 Docker Engine 的持久状态，由用户的 Builder/Cache 配置管理，Backend 不创建或删除它。

### R6 API defaults

- 管理 API 默认地址改为 `127.0.0.1:8080`。
- 默认不发送 `Access-Control-Allow-Origin: *`；跨域来源必须显式配置。

### R7 Tests and documentation

- 使用可注入 Docker 命令执行器验证 Context 解析、Run 参数、敏感信息屏蔽、Cleanup 和 CleanupAll，不依赖真实 GitHub 凭证。
- 增加配置、API 默认值、Backend 创建和跨层 Backend 枚举测试。
- 保持现有 Docker/Tart Backend 行为并运行项目规定的 `make unit-test`、`make check`、`prek run --all-files`。
- 更新 README、`config.example.yaml`、配置参考和 OrbStack 使用文档，禁止写入个人绝对路径或真实 Secret。

## Acceptance Criteria

- [x] `main` 未产生本次已跟踪修改，工作位于 `feat/docker-host-runner`。
- [x] `backend: docker-host` 配置验证通过，且第一阶段限制被验证。
- [x] Run 参数包含 `linux/amd64`、Socket 挂载和管理标签，不包含 `--privileged` 或 JIT 明文参数。
- [x] Runner 镜像固定版本、不会启动或包含 Dockerd，并包含 Docker CLI、Buildx、Compose。
- [x] Cleanup 使用 `rm -f -v`，只删除带精确管理标签的目标容器。
- [x] CleanupAll 只删除当前 Runner Set 的 EFR 管理容器。
- [x] 指定缓存目录跨 Job 保留，项目级 Cargo Target 不跨 Namespace 共享。
- [x] API 默认监听 `127.0.0.1:8080`，默认 CORS 不允许任意来源。
- [x] 新增测试全部通过，现有 Docker/Tart 回归结果已记录。
- [x] `make unit-test`、`make check`、`prek run --all-files` 的真实结果已记录。
- [x] 真实 PAT 只存在于仓库外的 `0600` 本地配置，未输出或提交；临时 Smoke 分支已删除，测试仓库 `main` 未修改。
- [x] Repository-level `max_runners: 1` 真实 JIT Job 成功，Runner 为 AMD64、通过 OrbStack Socket 使用宿主 Engine，且镜像中无 Dockerd。
- [x] Job 完成后 EFR Runner 容器和临时注册均为 0，缓存文件与命名卷保留，普通容器未受影响。

## Out of Scope

- CloudSpine Workflow 或其他生产 Workflow 接入。
- 多仓库、多 Job 并行、ARM64 Runner Set、Dashboard 重构、生产部署和远程唤醒。
- 删除或改变现有 DinD Docker Backend、Tart Backend。
- GitHub Release、GitHub App 配置或将真实 PAT 写入仓库。
