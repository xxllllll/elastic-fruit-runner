# CloudSpine Runner Image Readiness Implementation Plan

## 1. Preparation

- [x] 确认当前 HEAD 为 `feat/docker-host-runner` 的 `91f2563`，工作区只有已知的本地 Agent/Trellis 未跟踪目录。
- [x] 创建并切换到 `feat/cloudspine-runner-image-readiness`，基线保持 `feat/docker-host-runner`。
- [x] 后续仅显式暂存本任务文件，禁止 `git add -A`。

## 2. Backend Cache Tests First

- [x] 修改 `internal/backend/docker_host_test.go`，先将期望路径改为项目与 Platform 隔离的 `cargo-home`。
- [x] 增加 `/home/runner/.rustup` bind mount 和 `RUSTUP_HOME=/home/runner/.rustup` 参数断言。
- [x] 分别断言 AMD64、ARM64 的 Cargo Home、Rustup Home、Cargo Target、sccache 路径不同。
- [x] 运行定向测试并确认旧实现失败：

```bash
go test ./internal/backend -run 'TestDockerHost(RunArguments|RunArgumentsARM64|CachePlatformSegment)'
```

## 3. Backend Cache Implementation

- [x] 修改 `internal/backend/docker_host.go`：
  - 增加 `RUSTUP_HOME` 容器环境；
  - 把 Cargo Home 移到 `projectCacheRoot/cargo-home`；
  - 增加 `projectCacheRoot/rustup-home` 挂载；
  - 保持 pnpm store 与 Runner Tool Cache 共享；
  - 保持 BuildKit、Socket、标签、JIT 和清理行为不变。
- [x] 重新运行 Backend 定向测试并确认成功。

## 4. Runner Image Toolchain

- [x] 修改 `images/docker-host-runner/Dockerfile`：
  - 安装 `build-essential`、`cmake`、`pkg-config`；
  - 固定并校验 GitHub CLI `2.96.0` 的 AMD64、ARM64 资产；
  - 固定并校验 Rustup `1.29.0` 的 AMD64、ARM64 `rustup-init`；
  - 设置 `CARGO_HOME`、`RUSTUP_HOME` 和 Cargo bin PATH；
  - 保持非 root 用户、固定 Runner 版本和 daemon/runtime 删除逻辑。
- [x] 构建时对未知 `TARGETARCH` 返回包含实际架构值的错误。

## 5. Dual-Architecture Image Verification

- [x] 在 `images/docker-host-runner/` 增加双架构验证脚本。
- [x] 脚本验证基础工具、架构、PATH、Docker CLI 插件和 daemon/runtime 缺失状态。
- [x] 脚本为每个架构创建独立临时 Cargo/Rustup bind mount 目录，验证 Rustup 首次初始化与第二容器复用。
- [x] 脚本使用唯一资源名称和 trap 清理自身临时资源，不删除其他 Image、容器或 named volume。
- [x] 执行：

```bash
images/docker-host-runner/test.sh
```

## 6. Documentation

- [x] 更新 `images/docker-host-runner/README.md` 的工具列表、Rustup 初始化说明和双架构验证命令。
- [x] 更新 `doc-site/src/content/docs/how-to/docker-host-orbstack.md` 的缓存布局、Rustup Home、Platform 隔离和一次冷缓存说明。
- [x] 检查根 `README.md` 是否存在与新缓存布局冲突的表述，仅在确有冲突时修改。

## 7. Quality Verification

- [x] 运行格式和定向测试：

```bash
gofmt -w internal/backend/docker_host.go internal/backend/docker_host_test.go
go test ./internal/backend
git diff --check
```

- [x] 运行完整单元测试：

```bash
make unit-test
```

- [x] 在临时工具目录安装仓库要求的 golangci-lint `v2.11.3`，并通过 `uvx` 提供临时 `prek` 命令；不修改系统级依赖或仓库依赖文件。
- [x] 使用临时 PATH 执行完整检查：

```bash
PATH="<temporary-tools>:$PATH" make check
```

- [x] 再次执行双架构 Image 验证，保存实际版本与结果摘要。
- [x] 检查 `git status --short` 和 `git diff --stat`，确认没有 CloudSpine、凭证、根 CI 或本地 Agent 配置改动。

## 8. Final Contract Review

- [x] 使用 `trellis-check` 检查 PRD、设计、实现和代码一致性。
- [x] 更新 `.trellis/spec/backend/docker-host-integration.md`，将项目与 Platform 隔离的 Cargo/Rustup 缓存及 Image 工具契约记录为项目规范。
- [x] 重新运行受规范更新影响的文档检查和 `git diff --check`。
- [x] 完成前报告每条验收标准对应的命令和真实结果；无证据的项目不得标记完成。

## Rollback Points

- Backend 缓存测试未通过：只回退 `internal/backend/docker_host.go` 与对应测试，不触碰旧缓存目录。
- 任一架构 Image 构建失败：保留 Dockerfile 修改进行诊断，不开始 CloudSpine Smoke。
- Rustup 第二容器复用失败：任务保持未完成，检查 PATH、目录所有权和两个 Home 挂载，不退回全局共享 Cargo Home。
- 完整质量检查失败：只修复与本任务改动相关的问题；现存无关失败需记录证据，不得声明通过。
