# Docker Host ARM64 Runner Implementation Plan

## Implementation

1. 扩展 `config/docker_host.go` 的 Platform验证，允许 AMD64和 ARM64，保留
   AMD64默认值。
2. 更新配置测试：显式 ARM64成功，未知 Platform失败，原有限制继续生效。
3. 在 `internal/backend/docker_host.go` 增加受限 Platform到缓存目录 Segment
   的映射，并将项目级 Cargo Target与 sccache移入 Platform层级。
4. 更新 Backend测试，分别断言 AMD64/ARM64 `--platform` 参数和缓存目录，
   并测试未知 Platform不会创建目录或运行 Docker。
5. 修改 `images/docker-host-runner/Dockerfile`，按 `TARGETARCH` 选择 Compose
   Asset和 SHA-256，未知架构构建失败。
6. 更新 Runner Image README、`config.example.yaml`、OrbStack指南和配置参考，
   说明两个 Platform、独立本地 Tag和缓存迁移。
7. 构建并检查 AMD64镜像，确认现有行为没有回归。
8. 构建并运行 ARM64镜像，验证工具、Daemon删除、非 root用户及宿主 Socket。
9. 使用最小权限测试仓库执行一次 `max_runners: 1` ARM64真实 JIT Smoke。
10. 更新 Trellis Docker Host集成规范和任务验证证据。

## Validation commands

```sh
go test -count=1 ./config ./internal/backend ./internal/management ./internal/api

docker build \
  --platform linux/amd64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0-amd64 \
  images/docker-host-runner

docker build \
  --platform linux/arm64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0-arm64 \
  images/docker-host-runner

make unit-test
make check
prek run --all-files
cd doc-site && pnpm run build
git diff --check
```

## Live ARM64 smoke assertions

- GitHub Job `runs-on` 使用 ARM64 Runner Set名称。
- Runner报告 `uname -m = aarch64`。
- `docker version` 连接 OrbStack宿主 Engine。
- `docker run --rm --platform linux/arm64 alpine:3.22 uname -m` 返回
  `aarch64`。
- Runner镜像内 `dockerd`、`containerd`、`runc` 不存在。
- ARM64项目级缓存写入 `linux-arm64` 目录，AMD64目录不被修改。
- Job完成后 EFR管理容器和临时 Runner注册为 0。
- 命名卷、shared缓存和普通容器保留。

## Risk and rollback points

- Dockerfile架构映射错误会在镜像构建阶段通过 SHA-256或未知架构分支失败。
- 缓存布局变更只影响可重建的项目级缓存；不自动移动旧目录。
- 真实 Smoke前先完成无凭证 ARM64镜像和 Socket验证。
- 临时 Workflow只存在于临时测试分支，Smoke后删除，不修改测试仓库
  `main`。
- 不在本任务中放宽一个 Runner Set和 `max_runners: 1` 限制。

## Review gate before implementation

- PRD没有未解决问题。
- Design明确 Platform、镜像、缓存迁移、兼容性和回退。
- Implementation步骤覆盖代码、测试、双架构镜像、真实 Smoke和最终门禁。
- 用户批准规划后才运行 `task.py start` 并创建功能分支。

## Validation results

- Targeted Go tests passed for `config`, `internal/backend`,
  `internal/management`, and `internal/api`.
- AMD64 image `2.332.0-amd64` built as `linux/amd64`; ARM64 image
  `2.332.0-arm64` built as `linux/arm64` from the same pinned OCI Index.
- Compose `2.40.3` checksums passed for both `x86_64` and `aarch64`; an
  unsupported `TARGETARCH=riscv64` build failed with the expected diagnostic.
- Local ARM64 socket smoke passed: native `aarch64` Runner `2.332.0`, OrbStack
  Engine access, Buildx, Compose, ARM64 child container, no daemon binaries,
  and writes to the `linux-arm64` project cache paths.
- Real GitHub run `29325319753`, job `87060127209`, completed successfully on
  Runner `repo-orbstack-arm64-75017` with `max_runners: 1`.
- The live job verified runtime `Privileged=false`, EFR labels, ARM64 image,
  JIT environment presence without printing its value, all required mounts,
  native ARM64 child execution, and architecture-isolated cache writes.
- After the job, managed containers and repository runner registrations were
  zero; named volume, shared caches, AMD64 project cache, and unrelated
  containers were preserved.
- Controller shutdown deleted the listener session and ARM64 scale set. The
  temporary DockerDog branch was deleted and its `main` stayed at `fa137c2`.
- `make unit-test`, `make check`, `prek run --all-files`, golangci-lint, and
  the documentation site build passed.
