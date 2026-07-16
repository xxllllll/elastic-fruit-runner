# 支持 Docker Host ARM64 Runner

## Goal

在已经通过真实 Repository-level JIT Smoke 的 `docker-host` Backend 上增加
原生 Linux ARM64 Runner 支持，使 Apple Silicon + OrbStack 可以不经过 AMD64
模拟运行单个弹性 JIT Runner，同时保留现有 AMD64 行为。

## Background

- 基础分支为 `feat/docker-host-runner`，已完成并真实验证 AMD64 MVP。
- OrbStack Docker Engine 运行在 `linux/arm64`，并同时支持
  `linux/amd64` 与 `linux/arm64`。
- `ghcr.io/actions/actions-runner:2.332.0` 是多架构 OCI Index，固定 Index
  Digest 为 `sha256:8c3f5970b8ceb90cbd3e89b80c6806bb74d9c31686e9177c743323a4539d12f5`；
  其中 ARM64 Manifest Digest 为
  `sha256:5b922563ee9406d55e77a57d3dde85e1d793e600ed326ef6739182a847efae0f`。
- Docker Compose `2.40.3` 的 ARM64 Asset 为
  `docker-compose-linux-aarch64`，SHA-256 为
  `d26373b19e89160546d15407516cc59f453030d9bc5b43ba7faf16f7b4980137`。
- 当前配置验证只接受 `linux/amd64`，Runner Image Dockerfile 也主动拒绝
  非 AMD64 构建。
- 本阶段仍保持一个 Repository-level `docker-host` Runner Set 和
  `max_runners: 1`，不同时增加多个 Runner Set 或并发调度。

## Requirements

### R1 Platform support

- `docker-host` 接受 `platform: linux/amd64` 和 `platform: linux/arm64`。
- 未配置 `platform` 时继续默认 `linux/amd64`，避免改变现有配置行为。
- 其他 Platform 必须在配置验证阶段被拒绝，并返回包含实际值的错误。

### R2 Multi-architecture runner image

- `images/docker-host-runner/` 同一 Dockerfile 支持 AMD64 和 ARM64。
- GitHub Actions Runner 版本继续固定为 `2.332.0`，基础镜像继续固定 OCI
  Index Digest。
- Compose 版本继续固定为 `2.40.3`，根据 `TARGETARCH` 选择对应 Asset 和
  SHA-256；未知架构必须使镜像构建失败。
- 两个架构都不包含或启动 Dockerd、Containerd、Runc。
- 两个架构都包含 Docker CLI、Buildx 和 Compose，并继续使用非 root
  `runner` 用户执行 `/home/runner/run.sh`。

### R3 Cache isolation

- shared Cargo Home、pnpm store 和 Runner Tool Cache 继续跨架构共享。
- Cargo Target 与 sccache 必须同时按项目 Namespace 和 Runner Platform
  隔离，防止 AMD64 与 ARM64 产物相互污染。
- 项目级缓存统一迁移为
  `<cache_root>/<namespace>/<platform-segment>/{cargo-target,sccache}`；Platform
  Segment 固定映射为 `linux-amd64` 或 `linux-arm64`。
- 缓存路径不能包含未经验证的 Platform 字符串，也不能允许路径穿越。
- 现有 AMD64 项目级缓存允许一次冷启动；不自动复制或移动旧缓存，避免
  Backend 修改用户已有宿主目录。

### R4 Runtime and cleanup

- ARM64 Run 参数使用 `--platform linux/arm64`，继续挂载活动 Docker
  Context Unix Socket。
- 不使用 `--privileged`，不把 JIT Config 放入参数或日志。
- Cleanup、CleanupAll、管理标签和匿名卷删除语义不按架构分叉。
- Controller、GitHub Scale Set Client 和 Repository-level 生命周期契约不变。

### R5 Validation

- 增加 ARM64 配置、Run 参数、缓存路径和镜像架构测试。
- AMD64 现有测试必须继续通过。
- 分别构建或检查 AMD64、ARM64 Runner Image；本机至少真实运行 ARM64
  Image Smoke。
- 使用相同最小权限测试仓库执行一次 `max_runners: 1` 的真实 ARM64 JIT
  Job，验证原生架构、宿主 Docker Socket、缓存和 Job 后清理。
- 完成 `make unit-test`、`make check`、`prek run --all-files` 和文档站构建。

## Acceptance Criteria

- [x] `platform: linux/arm64` 配置验证通过，默认 Platform 仍为
      `linux/amd64`，其他值被拒绝。
- [x] ARM64 `docker run` 参数包含 `--platform linux/arm64`、Socket 和管理
      标签，不包含 `--privileged` 或完整 JIT Config。
- [x] 同一 Runner Dockerfile 可以构建 AMD64 和 ARM64，两个 Compose Asset
      都使用固定 SHA-256。
- [x] ARM64 镜像内 Runner、Docker CLI、Buildx、Compose 可用，Dockerd、
      Containerd 和 Runc 不存在。
- [x] 项目级 Cargo Target 与 sccache 在 AMD64/ARM64 之间隔离，共享缓存仍
      保持共享。
- [x] Cleanup 和 CleanupAll 回归通过，不删除命名卷、宿主缓存或用户容器。
- [x] 一次真实 ARM64 JIT Job 成功，Runner 与子容器均报告 `aarch64`，Job 后
      EFR 管理容器和临时 Runner 注册为 0。
- [x] 所有新增测试、现有回归和项目质量门禁通过。
- [x] 未提交真实 PAT、JIT Config、Token、个人 Socket 路径或测试仓库私有
      配置。

## Out of Scope

- 同时配置 AMD64 与 ARM64 两个 `docker-host` Runner Set。
- 多仓库、并发 Runner、组织级 Runner Set或 Dashboard 重构。
- 自动发布多架构镜像、GitHub Release、Homebrew Formula。
- CloudSpine、DockerDog 或其他生产 Workflow 接入。
- 将 Controller 安装为 macOS 常驻服务。
