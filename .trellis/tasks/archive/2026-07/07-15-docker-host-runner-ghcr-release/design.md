# Docker Host Runner GHCR Release Design

## 1. Delivery Sequence

本任务使用两个连续阶段，避免 GHCR 工作流与尚未进入 Fork `main` 的 Backend 代码相互依赖：

1. `feat/docker-host-runner` 通过 Fork PR 合并到 `main`。
2. 从最新 `main` 创建 `feat/docker-host-runner-ghcr-release`，实现并通过第二个 Fork PR。

不创建 `boring-design` 上游 PR。Trellis 文件保持本地，不进入产品提交。

## 2. Image Naming and Versioning

Repository：

```text
ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner
```

首个版本 Tag：

```text
2.332.0-r1
```

其中 `2.332.0` 必须由 Dockerfile 的 Actions Runner 基础镜像 Tag解析，`r1` 是同一 Runner
版本下的 Image 配方修订号。Workflow 只接受正整数 Revision，避免版本输入与 Dockerfile
漂移。

每次发布还生成：

```text
sha-<完整 Git commit SHA>
```

不发布 `latest`。两个 Tag 允许定位发布来源，但生产安全边界是 Registry 返回的 OCI Index
Digest，而不是 Tag 的不可变性。

## 3. Workflow Architecture

新增 `.github/workflows/docker-host-runner-image.yml`：

- `pull_request`：相关 Image、脚本或 Workflow 变化时，使用 hosted Ubuntu、QEMU 和 Buildx
  构建两个平台，但不登录、不推送。
- `workflow_dispatch`：只允许从默认分支发布；输入 `revision`，默认 `1`。
- 构建 Context 固定为 `images/docker-host-runner`。
- 发布 Job拥有 `contents: read` 和 `packages: write`；PR Job只有 `contents: read`。
- `docker/build-push-action` 一次构建 `linux/amd64,linux/arm64` 并推送单一 OCI Index。
- `provenance: false`、`sbom: false`，使 Index 平台项精确为两个可运行 Image。
- 使用 OCI labels 连接公开源码仓库、Source SHA 和 Image 版本。
- 发布后解析 Raw Index，验证平台集合、版本 Tag/SHA Tag Digest 一致，并把结果写入 Job
  Summary。

## 4. Verification Script

新增 `images/docker-host-runner/resolve-release.sh`：

1. 接收显式版本 Tag，不接受空值或 `latest`。
2. 使用 `docker buildx imagetools inspect` 获取远程 Raw OCI Index。
3. 验证 Media Type、两个且仅两个平台、架构无重复。
4. 获取版本 Tag 的 Index Digest。
5. 输出：

```text
image=<repository>@sha256:<index-digest>
amd64_digest=sha256:<digest>
arm64_digest=sha256:<digest>
```

脚本不修改配置，不读取 GitHub Token。Package 公开后可在匿名 Docker 环境执行。

已有 `test.sh` 增加可选远程 Image 模式，使同一工具契约既能构建本地 Image，也能验证
版本 Tag/Digest 下的两个平台，避免复制工具检查逻辑。

## 5. Host Digest Pinning

生产配置使用 Index Digest：

```yaml
image: ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:<index>
platform: linux/amd64
```

Docker 根据显式 Platform 从 Index 选择 AMD64 Manifest。更新时先解析并验证新版本，再备份
配置、替换单一 `image` 字段并重启 Controller。回退只需恢复旧 Digest 或配置备份，不删除
新旧 Image 和缓存。

## 6. Public Package Boundary

Runner Image 不包含仓库凭证。Package 设置为公开读取后，Controller 不需要保存 GHCR Token，
新主机和本地 Image 丢失场景均可按 Digest 拉取。发布权限仍只存在于 Fork Workflow 的短期
`GITHUB_TOKEN` 中。

## 7. Real Validation

发布后按以下顺序验证：

1. 未登录 Registry 的 Manifest 解析成功。
2. 生产配置固定 Index Digest；工具契约分别按 AMD64、ARM64 Manifest Digest 拉取，避免
   Docker Engine 在同一本地 Index 引用上拒绝第二个平台覆盖。
3. 使用 Digest 更新宿主配置并重启 LaunchAgent。
4. API 返回 Runner Set `repo-orbstack-amd64`，Image 为新 Digest。
5. 触发 CloudSpine `self-hosted-runner-smoke.yml@main`。
6. 等待成功，并检查 Runner 容器、GitHub临时注册、日志敏感信息与 Controller 连通状态。

## 8. Failure and Recovery

- Fork PR 检查失败：不合并，修复当前分支。
- 双架构构建失败：不发布 Tag，不修改宿主配置。
- Package 无法公开读取：不切换宿主配置。
- Manifest 平台集合错误或两个 Tag Digest 不一致：发布判定失败，不切换宿主。
- Controller 重启失败：恢复权限为 `0600` 的配置备份并重新 kickstart。
- JIT Smoke 失败：保留新 Image 证据进行诊断；根因属于新 Digest时恢复旧配置。

## 9. Implementation Outcome

- `r1` 的双架构 OCI Index 已成功发布，但发布后验证在同一本地 Index Digest 引用依次拉取
  两个平台时失败。Docker Engine 不允许第二个平台覆盖该本地 Digest 引用。
- 生产配置继续固定 OCI Index Digest；远程工具契约改为分别使用 AMD64、ARM64 Manifest
  Digest，保持双架构验证且不改变生产选择规则。
- 修复后发布 `2.332.0-r2`，Workflow、匿名 Registry 读取、两个平台工具契约和 CloudSpine
  两次真实 JIT Smoke 均成功。
