# Docker Host ARM64 Runner Design

## Architecture

本阶段扩展现有 `docker-host` Backend 的 Platform 能力，不改变 Controller、
GitHub Runner Scale Set Client 或 Backend 生命周期接口。

数据流保持：

`RunnerSetConfig.Platform -> DockerHostOptions.Platform -> docker run --platform -> OrbStack Engine`

支持的 Platform 集合固定为：

```text
linux/amd64
linux/arm64
```

未配置时继续默认 `linux/amd64`，保持现有配置兼容性。

## Configuration boundary

`config.validateDockerHostRunnerSet` 是 Platform 的入口验证位置。它接受两个
明确值，不允许任意 Docker Platform 字符串进入 Backend。

第一阶段其余限制不变：

- Repository-level only
- 全局最多一个 `docker-host` Runner Set
- `max_runners: 1`
- Unix Socket Docker Endpoint only

错误信息列出允许值和实际值，便于在不阅读源码时诊断配置。

## Runner image

基础镜像继续使用固定 OCI Index：

```dockerfile
FROM ghcr.io/actions/actions-runner:2.332.0@sha256:8c3f...
```

该 Index 同时包含 AMD64 和 ARM64 Manifest；Docker BuildKit 根据
`--platform` 选择对应 Manifest。无需维护两条 `FROM`。

Compose 安装根据 `TARGETARCH` 使用显式映射：

```text
amd64 -> docker-compose-linux-x86_64 -> AMD64 SHA-256
arm64 -> docker-compose-linux-aarch64 -> ARM64 SHA-256
other -> build failure
```

未知架构不使用默认 Asset，避免下载错误二进制后生成表面可构建但无法运行的
镜像。

镜像继续删除 Dockerd、Containerd、Ctr 和 Runc，可执行入口仍为
`/home/runner/run.sh`，用户仍为非 root `runner`。

## Cache layout and migration

共享缓存保持：

```text
cache_root/shared/cargo-home
cache_root/shared/pnpm-store
cache_root/shared/tool-cache
```

项目级缓存统一改为：

```text
cache_root/<namespace>/linux-amd64/cargo-target
cache_root/<namespace>/linux-amd64/sccache
cache_root/<namespace>/linux-arm64/cargo-target
cache_root/<namespace>/linux-arm64/sccache
```

Platform Segment 不由字符串替换生成，而是通过受限映射函数返回。即使未来
Backend 被绕过配置层直接创建，也不能将任意 Platform 变成宿主路径。

旧 AMD64 项目缓存不自动迁移。缓存属于可重建数据，自动移动会扩大 Backend
对用户宿主目录的写入范围，并可能与仍在运行的旧版本 Controller 冲突。

## Runtime and cleanup compatibility

`DockerHostBackend.runArgs` 已经直接读取 `options.Platform`，因此 ARM64 运行
参数不需要新的命令分支。Socket、附加组、JIT 环境变量、管理标签和缓存 Mount
规则完全复用。

Cleanup 以 EFR 管理标签、Runner Set 和 Runner Name识别资源，与 Platform
无关。ARM64 不增加新标签，也不改变 `docker rm -f -v` 行为。

## Local image tags

本地单架构构建使用明确 Tag，避免后一次构建覆盖前一次构建：

```text
elastic-fruit-runner/docker-host-runner:2.332.0-amd64
elastic-fruit-runner/docker-host-runner:2.332.0-arm64
```

未来发布多架构 OCI Index 时可以恢复统一版本 Tag，但该发布流程不属于本任务。

## Validation strategy

1. 配置单测验证默认 AMD64、显式 AMD64、显式 ARM64和拒绝其他值。
2. Backend 参数测试分别验证两个 Platform及架构隔离缓存路径。
3. 使用 Buildx 构建 AMD64 和 ARM64 单架构镜像。
4. ARM64 镜像在本机真实运行，验证 Runner 工具、无 Daemon 和 Socket访问。
5. 临时测试 Workflow 使用 ARM64 Runner Set执行一个真实 JIT Job。
6. Job 后检查管理容器、临时 Runner 注册、缓存、命名卷和普通容器。

## Rollback

- 配置继续使用 `linux/amd64` 即可保留原运行模式。
- 回退代码不会删除新 Platform缓存目录。
- ARM64本地镜像使用独立 Tag，不覆盖已验证 AMD64镜像。
- Cleanup资源格式没有变化，旧版本仍可按相同标签清理遗留容器。
