# Docker Host Runner GHCR Release Implementation Plan

## 1. Baseline and First Fork PR

- [x] 确认当前分支、远端、提交链、工作区和上游差异。
- [x] 使用临时固定版本工具完成 `make check` 与 Secret 扫描。
- [x] 创建 `feat/docker-host-runner -> main` Fork PR，检查差异后以 merge commit 合并。
- [x] 更新本地 `main` 并确认 Fork `main` 包含 Docker Host Backend 实现。

## 2. GHCR Branch

- [x] 从更新后的 `main` 创建 `feat/docker-host-runner-ghcr-release`。
- [x] 确认任务 `base_branch=main`、实现分支正确。

## 3. Tests First

- [x] 为发布版本解析、Tag 规则、双架构 Index 校验准备失败测试或 Shell fixture。
- [x] 先验证现有仓库缺少 GHCR Runner Image Workflow 和解析脚本。
- [x] 明确 PR 构建与手动发布的权限差异。

## 4. Implementation

- [x] 新增 GHCR 双架构 Buildx Workflow。
- [x] 新增远程版本解析与 Index 平台验证脚本。
- [x] 复用现有 Image 工具契约验证远程 Image。
- [x] 更新 Image README、根 README、配置示例和 OrbStack 指南。
- [x] 更新 Docker Host Trellis Spec 的发布与 Digest 契约。

## 5. Local Quality Verification

- [x] Shell syntax 和定向脚本测试通过。
- [x] Workflow YAML Parser 与 actionlint 通过。
- [x] `make check`、文档站构建、双架构本地 Image 契约通过。
- [x] `git diff --check`、敏感信息扫描、文件规模检查通过。

## 6. Second Fork PR

- [x] 提交并推送 GHCR 实现分支。
- [x] 创建只指向 Fork `main` 的 PR。
- [x] 检查 PR差异和可执行验证后合并。
- [x] 更新本地 `main`。

## 7. Real GHCR Release

- [x] 从 Fork `main` 手动触发 `revision=1` 发布并保留 `r1` 结果。
- [x] 修复 `r1` 发布后验证中同一本地 Index Digest 跨平台覆盖失败的问题。
- [x] 发布正式 `revision=2`，记录 Run、Source SHA、Tag 与 Digest。
- [x] 验证 Container Package 可公开读取。
- [x] 在匿名读取条件下验证 Index 和两个架构 Image。

## 8. Host Migration and JIT Smoke

- [x] 备份宿主配置并保持 `0600` 权限。
- [x] 只替换 Image 为 Index Digest，重启 LaunchAgent。
- [x] 验证 API、Listener、Runner Set、Image 和日志。
- [x] 触发并等待 CloudSpine 手动 Smoke 两次成功。
- [x] 验证 EFR 管理容器、GitHub临时 Runner、无关容器与缓存状态。

## 9. Final Records

- [x] 在 `validation.md` 记录所有真实命令、Run URL、Digest 和最终状态。
- [x] 运行 `trellis-check`，完成任务并归档。
- [x] 最终确认没有上游 PR、凭证或本地 Agent 配置进入提交。

## Validation Commands

```bash
PATH="<temporary-tools>:$PATH" make check
images/docker-host-runner/test.sh
bash -n images/docker-host-runner/*.sh
actionlint
pnpm --dir doc-site build
git diff --check
```

## Risky Files and Recovery Points

- `.github/workflows/docker-host-runner-image.yml`：发布权限和 Tag 生成错误会影响 Registry；发布前
  先完成 PR build。
- `/Users/CatServer/.elastic-fruit-runner/config.yaml`：只进行单字段替换，修改前保留 `0600`
  备份。
- LaunchAgent：只重启 `io.github.xxllllll.elastic-fruit-runner`，不终止其他进程。
- GHCR Package Visibility：只有双架构发布成功后才设置公开读取。
