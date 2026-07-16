# 发布 Docker Host Runner GHCR 双架构镜像

## Goal

先把已经通过真实 CloudSpine 生产验证的 `feat/docker-host-runner` 合并到
`xxllllll/elastic-fruit-runner:main`，再为 Docker Host Runner Image 建立可重复的
GHCR 双架构发布、OCI Index Digest 固定、宿主更新和真实 JIT 验证流程。

## Confirmed Facts

- 用户明确不向 `boring-design/elastic-fruit-runner` 提交上游 PR；所有 PR 只进入 Fork 的
  `main`。
- `feat/docker-host-runner` 已通过 Fork PR #1 合并到 `main`，Merge Commit 为
  `4d948192c3b7ab28a42cecf31ccde43103b0a2d6`。
- 当前 AMD64、ARM64 Runner Image 只有本地 Tag，`RepoDigests=[]`；按当前未限定域名的
  引用查询 Docker Hub 时返回 `denied/unauthorized`，不能作为可恢复的远程来源。
- Fork 是公开仓库，默认分支为 `main`，且 `main` 没有 Branch Protection。
- 当前 Dockerfile 基于固定 OCI Index Digest 的
  `ghcr.io/actions/actions-runner:2.332.0`，支持 BuildKit `TARGETARCH=amd64|arm64`。
- `images/docker-host-runner/test.sh` 已真实验证两个架构的工具契约和 Rustup 持久化。
- EFR 已有 GHCR 登录和 `packages: write` 的发布先例，但现有 Release Workflow 发布的是
  根 Controller Image，不是 `images/docker-host-runner`。
- CloudSpine 生产 Runner Set 使用 `platform: linux/amd64`，Controller 由 LaunchAgent
  `io.github.xxllllll.elastic-fruit-runner` 常驻运行。
- Image 发布使用 GitHub Actions 的仓库 `GITHUB_TOKEN`；正式 `r2` Package 已通过不读取
  本机 Docker 凭证的 Registry Bearer Token 请求确认可公开读取。

## Requirements

### R1. Fork 分支集成

- 创建 `feat/docker-host-runner -> main` 的 Fork PR，保留现有四个提交。
- 在合并前复核提交、敏感信息和本地质量证据；不得创建上游 PR。
- PR 合并后从最新 Fork `main` 创建 `feat/docker-host-runner-ghcr-release`。

### R2. GHCR Image Identity

- 发布地址固定为：
  `ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner`。
- 每次发布产生一个同时包含 `linux/amd64` 与 `linux/arm64` 的 OCI Index。
- 版本 Tag 使用 `<actions-runner-version>-r<image-revision>`，首版为 `2.332.0-r1`。
- 同时发布 `sha-<source-commit>` Tag；不得发布或依赖 `latest`。
- Tag 只用于发现，生产配置必须使用 Registry 返回的 OCI Index Digest。

### R3. GitHub Actions Publishing

- 新增手动 `workflow_dispatch` 发布入口，并为相关 Pull Request 提供不推送的双架构构建。
- 发布 Job 仅使用 `contents: read` 与 `packages: write`，不得引入 PAT 或长期 Package Token。
- 使用 Buildx/QEMU 构建并推送双架构 OCI Index，关闭额外 attestation manifest，使平台集合
  精确为 AMD64、ARM64 两项。
- Image 必须包含 OCI source、revision、version 标注，并继续执行已有工具契约验证。
- 发布完成后验证版本 Tag 与 SHA Tag解析到同一个 Index Digest。

### R4. Public Pull and Digest Verification

- GHCR Package 设置为公开读取；匿名环境必须能检查 Manifest 并拉取 AMD64 Image。
- 验证 Index 仅包含 `linux/amd64` 和 `linux/arm64`，并记录 Index Digest 与两个平台 Digest。
- Digest 引用必须能执行工具契约、非 root 用户检查和最小 Docker Socket 探针。

### R5. Image Update Contract

- 文档描述版本规则、发布步骤、Digest解析、宿主更新、验证和回退。
- 提供仓库内脚本，从版本 Tag解析并验证双架构 Index，输出适合配置的 Digest 引用。
- 脚本不得写宿主配置、读取凭证或静默选择 `latest`。
- 后续更新必须先发布新版本，再显式替换宿主 Digest；旧 Digest 保持可恢复。

### R6. Host Migration

- 对 `/Users/CatServer/.elastic-fruit-runner/config.yaml` 创建权限保持为 `0600` 的备份。
- 只替换 CloudSpine Runner Set 的 `image` 字段为 GHCR Index Digest 引用；不得输出 PAT。
- 保持 `platform: linux/amd64`、Runner Set、缓存目录、API 和并发限制不变。
- 重启 LaunchAgent 后验证 Controller API、Listener 和空闲资源状态。

### R7. Real JIT Validation

- 从 GHCR Digest 引用创建真实 CloudSpine JIT Runner。
- 手动触发已有 `self-hosted-runner-smoke.yml`，并等待最终成功。
- 证明实际 Runner 使用新 GHCR Digest、AMD64 Platform、宿主 ARM64 Docker Socket和现有缓存。
- Job 完成后 EFR 管理容器与 GitHub 临时 Runner 注册均为 0。
- Controller 日志不得包含 PAT、GitHub Token 或 JIT Config 实值。

## Acceptance Criteria

- [x] Fork PR 已合并进 `xxllllll/elastic-fruit-runner:main`，没有上游 PR。
- [x] GHCR 发布实现及后续修复通过独立 Fork PR 合并进 `main`。
- [x] GHCR Package 可匿名读取。
- [x] 正式版 `2.332.0-r2` 与 SHA Tag 指向同一个 OCI Index Digest；`r1` 保留并记录其
  发布后跨平台本地拉取验证失败原因。
- [x] OCI Index 平台集合精确为 `linux/amd64`、`linux/arm64`。
- [x] 仓库脚本可以由版本 Tag输出并验证固定 Digest 引用。
- [x] AMD64 与 ARM64 远程 Image 工具契约通过。
- [x] `make check`、文档构建、Workflow YAML/actionlint 和 `git diff --check` 通过。
- [x] 宿主配置使用 GHCR Index Digest，原配置备份权限为 `0600`。
- [x] LaunchAgent 重启后 Controller 正常监听 `127.0.0.1:38080`。
- [x] CloudSpine 真实 JIT Smoke 连续两次成功，运行后无 Runner 容器和临时注册残留。
- [x] 代码、提交、Workflow、日志和任务记录中没有凭证实值。

## Out of Scope

- 向 `boring-design/elastic-fruit-runner` 创建 PR。
- 为其他仓库发布或接管通用 EFR Controller Image。
- 多仓库、多 Runner Set、并发 Runner 或缓存架构重构。
- 修改 CloudSpine Workflow 行为；本任务只触发已有手动 Smoke。
- 删除旧本地 Image、旧缓存或无关工作区文件。
