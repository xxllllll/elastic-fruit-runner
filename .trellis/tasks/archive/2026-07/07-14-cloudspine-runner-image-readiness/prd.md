# 准备 CloudSpine Runner Image 工具链

## Goal

使 EFR `docker-host` AMD64 Runner 能执行 CloudSpine 的只读真实集成 Smoke，且保持 Runner Image 通用、双架构一致，不绑定 CloudSpine 的 Rust、Node 或 pnpm 版本。

## Background and Confirmed Facts

- 当前任务基于 `feat/docker-host-runner`，实现分支为 `feat/cloudspine-runner-image-readiness`。
- 当前 Image 基于 Ubuntu 24.04，已有 Git、curl、jq、Python 3、Docker CLI、Buildx、Compose、OpenSSL、timeout；缺少 `gh`、Rustup/Cargo、Node/npm/pnpm、C/C++ 构建工具、`pkg-config` 与 CMake。证据：`images/docker-host-runner/Dockerfile:1-46` 及本地 AMD64/ARM64 容器命令检查。
- CloudSpine 生产 Workflow 直接调用 `gh api`、`gh pr`、`gh workflow`、`gh release`，因此 `gh` 是 Runner 基础依赖，不只是 Smoke 工具。
- CloudSpine 通过 `actions/setup-node@v6` 安装 `.node-version` 指定的 Node 24，并从 `package.json` 安装 `pnpm@11.12.0`；EFR Image 不应预装项目版本的 Node 或 pnpm。
- CloudSpine `rust-toolchain.toml` 固定 Rust `1.97.0`，当前 Workflow 直接执行 `rustup show active-toolchain`；EFR 应提供通用 Rustup 启动能力，但不预装 Rust `1.97.0`。
- CloudSpine 依赖 `aws-lc-sys`，其构建调用系统 CMake；因此最低系统构建工具不只是 `build-essential` 与 `pkg-config`，还需要 `cmake`。
- Backend 当前把整个 `/home/runner/.cargo` 挂载到 `shared/cargo-home`，并设置 `CARGO_HOME`，但没有设置或持久化 `RUSTUP_HOME`。证据：`internal/backend/docker_host.go:117-139,207-228`。
- Rustup 代理程序位于 `$CARGO_HOME/bin`，工具链位于 `$RUSTUP_HOME`（默认 `~/.rustup`）。只持久化 Cargo Home 不能复用已安装的 Rust 工具链。
- 当前 `shared/cargo-home` 跨 AMD64、ARM64 和项目共享；一旦 Rustup 安装代理程序，该目录会包含目标架构的可执行文件，不能继续安全跨架构共享。
- 用户已确认将 Cargo Home 迁移到项目与 Platform 隔离目录，并新增同级 Rustup Home；允许首次运行出现一次冷缓存。
- BuildKit 状态由宿主 Docker Engine 保存，不需要挂载进 Runner 容器；Runner 清理不会删除宿主 BuildKit 状态。证据：`doc-site/src/content/docs/how-to/docker-host-orbstack.md:76-95,136-141`。
- 当前仓库已有单元测试和 GitHub Actions，但没有 Runner Image 工具契约测试。

## Requirements

### R1. 通用系统工具链

- Runner Image 必须在 AMD64 与 ARM64 上提供 `gh`、`make`、`gcc`、`g++`、`pkg-config`、`cmake`。
- `gh` 必须使用固定版本和架构对应校验值安装，避免依赖构建时变化的系统仓库版本。
- 不恢复 `dockerd`、containerd 或 `runc`，不增加 `--privileged`，不改变宿主 Docker Socket 模型。

### R2. 通用 Rustup 启动能力

- Image 必须将 `/home/runner/.cargo/bin` 加入 `PATH`。
- Image 必须提供固定版本、校验值验证的 Rustup 初始化程序，但不得预装具体 Rust toolchain。
- 首次 Job 必须能初始化持久化 Rustup；后续由项目的 `rust-toolchain.toml` 选择工具链。

### R3. Rust 缓存正确性

- EFR 必须持久化 `RUSTUP_HOME`，否则每个 JIT Runner 都会重新下载 Rust 工具链。
- Cargo Home 与 Rustup Home 必须按项目和 Platform 隔离，不能让项目或架构共享 Cargo 配置、架构相关二进制或 toolchain。
- 缓存目录变更必须有单元测试、文档说明和一次冷缓存迁移提示。

### R4. 双架构 Image 契约验证

- 增加可重复执行的 Image 验证入口，分别构建并检查 AMD64、ARM64 Image。
- 验证必须覆盖架构、基础工具、`gh`、C/C++ 工具链、CMake、`rustup-init`、Cargo PATH、Docker CLI/Buildx/Compose，并确认 `dockerd` 仍不存在。
- 至少验证一次空缓存 Rustup 初始化，证明持久化目录中可生成并执行当前架构的 `rustup` 代理程序。

### R5. EFR 质量门禁

- 完成 `make unit-test`。
- 完成 `make check`；使用仓库配置对应的 golangci-lint v2.11.x 与 prek。
- 完成双架构 Image 构建及工具契约测试。
- 实现阶段只显式暂存任务文件，禁止使用 `git add -A`，避免误提交 `.agents/`、`.claude/`、`.codex/`、`.trellis/` 等本地开发目录。

## Acceptance Criteria

- [x] AMD64 与 ARM64 Image 中 `gh`、`make`、`gcc`、`g++`、`pkg-config`、`cmake`、`rustup-init` 均可执行。
- [x] Image 的 `PATH` 包含 `/home/runner/.cargo/bin`，但 Image 本身不包含 CloudSpine 固定版本的 Rust、Node 或 pnpm。
- [x] Backend 设置并挂载持久化 `RUSTUP_HOME`。
- [x] Cargo Home 与 Rustup Home 不再跨 AMD64、ARM64 共用同一目录。
- [x] 空缓存执行 Rustup 初始化成功，第二个同架构容器可直接执行缓存中的 `rustup`。
- [x] AMD64 与 ARM64 不会互相执行对方缓存中的 Rustup/Cargo 二进制。
- [x] `docker`、`docker buildx`、`docker compose` 保持可用，`dockerd` 保持不可用。
- [x] `make unit-test`、`make check` 与双架构 Image 验证均有真实成功输出。
- [x] README 与 docker-host 文档准确描述新增工具、Rustup 初始化和新缓存布局。
- [x] 未修改 CloudSpine 仓库，未写入 PAT、JIT 配置或其他凭证。

## Out of Scope

- 修改 CloudSpine Workflow、Trellis 任务或生产发布流程。
- 预装 CloudSpine 的 Rust 1.97.0、Node 24 或 pnpm 11.12.0。
- 多仓库、多 Runner Set、并发 Runner 或 LaunchAgent。
- 永久切换 CloudSpine 的 `runs-on`、移除 `Swatinem/rust-cache`、修改镜像 `platforms`。
- 修改 EFR 根级 GitHub Actions CI；该项需要单独确认。
- 提交本地 Agent/Trellis 配置目录。
