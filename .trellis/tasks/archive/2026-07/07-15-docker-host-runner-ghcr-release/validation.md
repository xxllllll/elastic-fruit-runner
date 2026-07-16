# Docker Host Runner GHCR Release Validation

## 1. Fork Pull Requests

- PR #1: <https://github.com/xxllllll/elastic-fruit-runner/pull/1>
  - Result: merged
  - Merge Commit: `4d948192c3b7ab28a42cecf31ccde43103b0a2d6`
  - Scope: Docker Host Backend、OrbStack Unix Socket、AMD64/ARM64、缓存、通用 Runner Image
    与 CloudSpine 工具链。
- PR #2: <https://github.com/xxllllll/elastic-fruit-runner/pull/2>
  - Result: merged
  - Merge Commit: `4ab366ef18007626b58fa3a16839f75938bc1452`
  - Implementation Commit: `a3f4049 feat(docker-host): 增加 GHCR 双架构发布`
- PR #3: <https://github.com/xxllllll/elastic-fruit-runner/pull/3>
  - Result: merged
  - Merge Commit: `d4a7987bd014b49cf1c06a774f14d687ab9cb29f`
  - Fix Commit: `3426f5d fix(docker-host): 按平台验证远程镜像`
- PR #4: <https://github.com/xxllllll/elastic-fruit-runner/pull/4>
  - Result: merged
  - Merge Commit: `0f1ee0fe703acc630ce9e5c8cbb5979334e3d531`
  - Fix Commit: `3a11457 fix(docker-host): 固定 GHCR 镜像摘要`

最终 Fork `main` 为 `0f1ee0fe703acc630ce9e5c8cbb5979334e3d531`。Fork 当前没有开放
PR；`boring-design/elastic-fruit-runner` 当前没有来自 `xxllllll` Fork 的开放 PR。

## 2. Pull Request Contract Runs

- PR #2 Hosted Runner Contract:
  - Run: <https://github.com/xxllllll/elastic-fruit-runner/actions/runs/29396432354>
  - Result: success
  - Job: `87290881911`
  - Duration: `2m39s`
- PR #3 Hosted Runner Contract:
  - Run: <https://github.com/xxllllll/elastic-fruit-runner/actions/runs/29397497339>
  - Result: success
  - Job: `87294213460`
  - Duration: `2m40s`

## 3. GHCR Releases

Image Repository:

```text
ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner
```

### r1

- Run: <https://github.com/xxllllll/elastic-fruit-runner/actions/runs/29396756599>
- Result: failure after the build and push had succeeded.
- Index Digest: `sha256:1fdbc3d1aaffccd1a3b0cbeda019f7fca893a05f3fbcc018366e0179f03b90c3`
- AMD64 Manifest: `sha256:bdf27f039f3168f38465f7b453a13a0dc0c6b6e31eb2f9949b2999f1056cb428`
- ARM64 Manifest: `sha256:e552cb665397030de1025e1967847478260dd6c0ec9076f04a9def72e873463b`
- Failure cause: 远程契约先按 Index Digest 拉取 AMD64，再用相同本地 Index Digest 引用拉取
  ARM64，Docker Engine 返回 `cannot overwrite digest`。`r1` 没有被覆盖或删除。

### r2

- Run: <https://github.com/xxllllll/elastic-fruit-runner/actions/runs/29397734937>
- Result: success
- Job: `87294953293`
- Duration: `3m9s`
- Source SHA: `d4a7987bd014b49cf1c06a774f14d687ab9cb29f`
- Tags:
  - `2.332.0-r2`
  - `sha-d4a7987bd014b49cf1c06a774f14d687ab9cb29f`
- 两个 Tag 指向同一 OCI Index Digest：
  `sha256:24d7af1adc02c8c5d21306752d3d31df1d693eeea0c9c59be4c3f481dc9911a8`
- Linux AMD64 Manifest:
  `sha256:f2d8767df95313a6f44f0a46e809498e08f6db861b387adf17ac7953af850ec2`
- Linux ARM64 Manifest:
  `sha256:b15b32772d27cca7a06e44b2f6ec4248803351cdd9310a812834e865031432a1`

生产引用：

```text
ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:24d7af1adc02c8c5d21306752d3d31df1d693eeea0c9c59be4c3f481dc9911a8
```

不读取本机 Docker 凭证的 Registry Bearer Token 请求返回 `HTTP/2 200`，Media Type 为
`application/vnd.oci.image.index.v1+json`，`docker-content-digest` 与上述 Index Digest
一致。

## 4. Local Quality Evidence

以下检查已通过：

```text
make check
images/docker-host-runner/test.sh
images/docker-host-runner/resolve-release_test.sh
bash -n images/docker-host-runner/*.sh
actionlint .github/workflows/docker-host-runner-image.yml
pnpm --dir doc-site build
git diff --check
```

`make check` 覆盖 Dashboard build、`go vet ./...`、Go build、`golangci-lint v2.11.3`、
`prek 0.4.9`、`goimports v0.42.0` 和全部 Go 单元测试。AMD64、ARM64 本地及 GHCR 远程
工具契约均通过。

## 5. Host Migration

- Config: `/Users/CatServer/.elastic-fruit-runner/config.yaml`
- Config Backup:
  `/Users/CatServer/.elastic-fruit-runner/config.yaml.backup-20260715-154453-pre-ghcr-digest`
- Config 与备份权限均为 `0600`。
- 仅 Runner Set 的 `image` 字段替换为 `r2` Index Digest；`platform: linux/amd64`、
  `max_runners: 1`、`cache_namespace: cloudspine`、`api_addr: 127.0.0.1:38080` 和
  `idle_timeout: 15m` 保持不变。
- Controller Binary Backup:
  `/Users/CatServer/.local/bin/elastic-fruit-runner.backup-20260715-154453-pre-ghcr-digest`
- Final Binary SHA-256:
  `3c6afe5ff2e7fa825a2b4aa4e83d9cdd9788a986972628b175c97321f3af80d8`
- Final Build Revision: `0f1ee0fe703acc630ce9e5c8cbb5979334e3d531`
- `vcs.modified=false`

LaunchAgent `io.github.xxllllll.elastic-fruit-runner` 重启后 PID 从 `11171` 变为 `14692`。
Controller API 返回 HTTP 200，监听 `127.0.0.1:38080`，Runner Set 返回固定 GHCR Index
Digest。最终空闲状态为 `managed_containers=0`、`registered_runners=0`。

## 6. CloudSpine Real JIT Smoke

- Workflow Run: <https://github.com/xxllllll/CloudSpine/actions/runs/29398544138>
- CloudSpine Source: `ce3549865e977bba2d01570a84e7e62c6533296b`

### Attempt 1

- Result: success
- Job: `87297530134`
- Runner: `repo-orbstack-amd64-37512`
- Start: `2026-07-15T07:48:01Z`
- End: `2026-07-15T08:04:34Z`

验证通过：x86_64 Runner、ARM64 OrbStack Docker Server、`linux/amd64` 子容器、Rust
1.97.0、Node 24.18.0、pnpm、仓库验证、Compose、Runtime/Backup AMD64 Image 构建及
6/6 目录缓存。BuildKit 本次为 cold。

### Attempt 2

- Result: success
- Job: `87301040200`
- Runner: `repo-orbstack-amd64-d3b76`
- Start: `2026-07-15T08:07:24Z`
- End: `2026-07-15T08:10:28Z`

热缓存证据：

```text
warm: /home/runner/.cargo
warm: /home/runner/.rustup
warm: /home/runner/.cache/efr/cargo-target
warm: /home/runner/.cache/sccache
warm: /home/runner/.cache/pnpm-store
warm: /opt/hostedtoolcache
BuildKit cache: warm
Directory caches warm at start: 6/6
```

两个 Attempt 实际 Runner 容器均使用同一固定 GHCR Index Digest。Attempt 2 完成后：

```text
managed_containers=0
registered_runners=0
preexisting_containers_missing=0
preexisting_volumes_missing=0
```

Controller 保持 HTTP 200；日志扫描未发现 PAT、GitHub Token 或 JIT Config 实值。

## 7. Final Remote Audit

- 本地分支：`main...origin/main`
- Fork `main`：`0f1ee0fe703acc630ce9e5c8cbb5979334e3d531`
- Fork Open PR: `0`
- Upstream Open PR from `xxllllll`: `0`
- 本任务产生且依赖不存在自托管 Runner 的 3 个排队 Run 已取消：
  - `29398255324` Unit Test
  - `29398255351` Integration Test
  - `29398255323` golangci-lint
- 未取消成功的发布与 Hosted Runner Contract Run。
- 工作区产品代码无未提交修改；仅 `.agents/`、`.claude/`、`.codex/`、`.trellis/` 为本地
  未跟踪目录，未进入 Fork 提交。
