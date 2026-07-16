# Docker Host Runner MVP Implementation Plan

## Implementation

- [x] 扩展配置：`cache_root`、`cache_namespace`、`docker-host` 第一阶段验证、API 默认值。
- [x] 新增可注入 Docker CLI 执行器和 Unix Endpoint Resolver。
- [x] 实现 `DockerHostBackend.Run`、缓存目录与挂载、标签和敏感信息屏蔽。
- [x] 实现标签限定的 `Cleanup`、`CleanupAll` 与 `rm -f -v`。
- [x] 在 Management、Proto、API、Dashboard Backend 映射中注册 `docker-host`。
- [x] 新增固定版本 Runner Image，移除 Daemon、添加 Compose、设置 Runner 入口。
- [x] 更新 README、配置示例、配置参考和 OrbStack 使用文档。
- [x] 复核并修复现有 Tart 基线测试的 10ms 进程启动时序不稳定问题。

## Validation

1. `go test -count=1 ./config ./internal/backend ./internal/management ./internal/api`
2. `cd dashboard && pnpm run build`
3. `docker build --platform linux/amd64 -t elastic-fruit-runner/docker-host-runner:2.332.0 images/docker-host-runner`
4. 以覆盖入口的无凭证容器检查镜像工具、无 Dockerd、Socket 权限与 Docker Engine 连通性。
5. `make unit-test`
6. `make check`
7. `prek run --all-files`
8. `git diff --check`、`git status --short --branch`、Secret 字符串审查。

## Risk and rollback points

- Docker Context 解析和 Cleanup 标签逻辑先由 Fake Command 单测确认，再执行本机无凭证 Docker Smoke。
- Socket 权限依赖 OrbStack 映射行为；验证 Runner 用户加入附加组后能调用 `docker version`。
- Proto 变更后使用 `buf generate`，并检查生成文件和 Dashboard 映射同步。
- 不运行任何需要 GitHub JIT Config、PAT 或 GitHub App Key 的命令。

## Validation results

- `make unit-test`: passed.
- `make check`: passed with temporary local installations of CI-pinned
  `golangci-lint v2.11.3`, `goimports v0.42.0`, and `prek 0.4.9` because the
  executables were not preinstalled on the host.
- `prek run --all-files`: passed.
- Dashboard and documentation site builds: passed.
- AMD64 runner image build and no-credential socket smoke: passed.
- Real local cleanup smoke: runner container and anonymous volume removed;
  named volume and unrelated container preserved; no EFR smoke containers left.
- Real repository-level JIT smoke: GitHub run `29321785751` and job
  `87048562819` completed successfully with Runner `2.332.0` and
  `repo-orbstack-amd64-a6a47`.
- The live job reported `x86_64`, reached the OrbStack host Engine
  (`server_arch=arm64`), ran an AMD64 Alpine child container, found no Dockerd,
  and verified Docker Buildx and Compose.
- All five configured cache mounts persisted the run stamp after cleanup;
  managed runner containers and repository runner registrations returned to
  zero. The named smoke volume and all pre-existing ordinary containers were
  preserved.
- Controller shutdown deleted the listener session and temporary runner scale
  set. The temporary DockerDog branch was deleted and its `main` branch stayed
  at `fa137c2`.
