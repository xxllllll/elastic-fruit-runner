# CloudSpine Runner Image Readiness Design

## 1. Scope and Boundaries

本任务在 EFR 内完成两个相互依赖的能力：

1. 通用 `docker-host` Runner Image 提供 CloudSpine 编译所需的系统基础工具和 Rustup 初始化入口。
2. `docker-host` Backend 为 Cargo 与 Rustup 提供项目、Platform 隔离的持久化目录。

Image 与缓存必须同时修改。只修改 Image 会导致 Rust toolchain 每个 Job 重新下载，并可能让 AMD64、ARM64 共用错误架构的 Cargo/Rustup 二进制。

任务不拆分为子任务：Dockerfile、运行时环境、缓存挂载和双架构验证构成一个完整运行契约，分别实现会产生不可独立验收的中间状态。

## 2. Runner Image Design

### 2.1 System packages

通过 Ubuntu 包安装以下通用构建组件：

- `build-essential`
- `cmake`
- `pkg-config`

`build-essential` 提供 GCC、G++、Make 和 binutils；CMake 单独安装，用于 CloudSpine 当前 `aws-lc-sys` 依赖。安装使用 `--no-install-recommends`，完成后删除 apt lists。

### 2.2 GitHub CLI

不依赖 Ubuntu 仓库中的滚动版本。Dockerfile 固定 GitHub CLI 版本，按 `TARGETARCH` 选择官方 `linux_amd64` 或 `linux_arm64` tarball，并使用架构对应 SHA-256 验证后安装到 `/usr/local/bin/gh`。

规划时核验的版本为 GitHub CLI `2.96.0`。实现时保持版本、下载 URL、两个架构校验值在 Dockerfile 同一位置，避免版本与校验值分散。

### 2.3 Rustup bootstrap

Dockerfile 固定 Rustup `1.29.0`，按 `TARGETARCH` 映射到：

- AMD64：`x86_64-unknown-linux-gnu`
- ARM64：`aarch64-unknown-linux-gnu`

从 Rust 官方版本归档下载 `rustup-init`，验证架构对应 SHA-256 后安装到 `/usr/local/bin/rustup-init`。

Image 不执行 Rustup 初始化，不包含具体 toolchain。运行 Job 首次调用：

```bash
rustup-init -y --profile minimal --default-toolchain none --no-modify-path
```

随后项目通过 `rust-toolchain.toml` 触发 Rust `1.97.0` 安装。EFR 不解析项目 toolchain 文件。

### 2.4 Runtime environment

Image 提供默认环境：

```text
CARGO_HOME=/home/runner/.cargo
RUSTUP_HOME=/home/runner/.rustup
PATH=/home/runner/.cargo/bin:<existing PATH>
```

Backend 在 `docker run` 参数中显式传递同样的 `CARGO_HOME` 与 `RUSTUP_HOME`，使运行契约不依赖调用方覆盖 Image 环境。

Node 与 pnpm 继续由 Workflow 安装。Image 初始状态下 `node`、`npm`、`pnpm`、`cargo`、`rustup` 均不要求存在；只有 `rustup-init` 是基础能力。

## 3. Cache Layout Design

### 3.1 New layout

```text
cache_root/
  shared/
    pnpm-store/
    tool-cache/
  <cache_namespace>/
    linux-amd64/
      cargo-home/
      rustup-home/
      cargo-target/
      sccache/
    linux-arm64/
      cargo-home/
      rustup-home/
      cargo-target/
      sccache/
```

挂载关系：

| Host directory | Container directory |
|---|---|
| `<namespace>/<platform>/cargo-home` | `/home/runner/.cargo` |
| `<namespace>/<platform>/rustup-home` | `/home/runner/.rustup` |
| `<namespace>/<platform>/cargo-target` | `/home/runner/.cache/efr/cargo-target` |
| `<namespace>/<platform>/sccache` | `/home/runner/.cache/sccache` |
| `shared/pnpm-store` | `/home/runner/.cache/pnpm-store` |
| `shared/tool-cache` | `/opt/hostedtoolcache` |

Cargo Home 采用项目与 Platform 双重隔离，原因包括：

- `$CARGO_HOME/bin` 含架构相关的 Rustup/Cargo 代理程序；
- Cargo Home 可能包含项目私有 registry 配置或凭证；
- 不同项目可以独立清理缓存，不影响其他 Runner Set。

Rustup Home 同样项目与 Platform 隔离，避免 toolchain 架构冲突。

### 3.2 Existing caches

旧 `shared/cargo-home` 不自动复制、不自动删除。自动复制可能把错误架构二进制或项目配置带入新目录；自动删除不符合安全规则。

首次使用新版本时 Cargo registry、Rustup 和目标产物均可能重新下载或编译。文档明确记录一次冷缓存行为。

### 3.3 BuildKit

BuildKit 状态继续由宿主 Docker Engine 管理，不增加容器 bind mount。EFR Runner 清理只删除带 EFR 管理标签的 Runner 容器，不修改 BuildKit 状态。

## 4. Image Contract Verification

在 `images/docker-host-runner/` 增加可直接执行的双架构验证脚本，不修改根 Makefile或 GitHub Actions。

脚本对 AMD64、ARM64 分别执行：

1. 使用对应 `--platform` 构建临时 Image。
2. 核验 `uname -m` 与目标架构一致。
3. 核验 `gh`、GCC、G++、Make、pkg-config、CMake、Docker CLI、Buildx、Compose、`rustup-init`。
4. 核验 `/home/runner/.cargo/bin` 已进入 `PATH`。
5. 核验 `dockerd`、containerd、runc 仍不存在。
6. 使用该架构独立的临时目录，通过 bind mount 挂载 Cargo Home 与 Rustup Home。
7. 第一个容器运行最小 Rustup 初始化。
8. 第二个同架构容器直接执行缓存中的 `rustup --version`，证明跨 JIT 容器复用成立。
9. 使用 trap 删除脚本创建的临时 Image 与临时目录，不接触其他容器或卷。

Backend 单元测试负责验证 AMD64、ARM64 生成不同的 Cargo/Rustup Host 路径；Image 脚本负责验证实际二进制和持久化行为。

## 5. Compatibility and Rollback

- 配置文件格式不变；现有 `cache_root`、`cache_namespace` 和 `platform` 字段继续使用。
- 新版本首次运行会创建新缓存目录，不迁移旧 Cargo Home。
- 回滚到旧提交后，Backend 会重新使用旧 `shared/cargo-home`，新目录保持未使用状态，不造成数据丢失。
- Runner Image 继续使用非 root `runner` 用户，不增加凭证、宿主路径或 CloudSpine 专用版本。
- Docker Socket、JIT 环境传递、管理标签及清理选择条件不改变。

## 6. Documentation and Spec Updates

实现同步更新：

- `images/docker-host-runner/README.md`：新增工具、Rustup 初始化和验证命令。
- `doc-site/src/content/docs/how-to/docker-host-orbstack.md`：更新缓存布局和冷缓存说明。
- `.trellis/spec/backend/docker-host-integration.md`：在完成验证后记录新的 Cargo/Rustup 缓存契约。

根级 CI、CloudSpine Workflow 和 LaunchAgent 不在本任务内。
