# Docker Host Runner MVP Design

## Architecture

配置从 `config` 进入 `internal/management.createBackend`，转换为 `backend.DockerHostOptions`。Controller 继续依赖现有 `backend.Backend`，无需改变 Scale Set 调度接口。

数据流：

`RunnerSetConfig` -> `DockerHostBackend` -> Docker Context Resolver -> Docker CLI -> OrbStack Engine -> JIT Runner Container

生命周期：

1. Listener 提供期望 Runner 数量。
2. Controller 生成一次性 JIT Config，调用 `Run`。
3. Backend 解析 Unix Socket、创建缓存目录并 detached 启动容器。
4. Controller 按现有 Backend 契约将 Runner 标记为空闲；GitHub 的 `JobStarted` 事件确认实际接单。
5. JIT Runner 完成一个 Job 后退出，`JobCompleted` 事件使 Controller 异步调用 `Cleanup`。
6. Controller 启动时调用 `CleanupAll`，依靠管理标签清理遗留容器。

## Backend contract

现有接口足以表达第一阶段生命周期。当前 Docker Backend 的 `Run` 在 `docker run -d` 成功后立即返回；它没有验证 Runner 已在 GitHub Online。Controller 依赖 Backend 返回值并将 Runner 标记为空闲，实际接单由 `JobStarted` 确认。第一阶段保持该调度模型，但新 Backend 至少验证 Docker 创建命令成功，不扩大 Controller 或 GitHub Client 契约。

## Docker CLI boundary

- `dockerCommandRunner` 统一执行 Docker CLI，并允许向子进程注入环境变量。
- `DOCKER_HOST` 非空时直接解析；否则显式读取当前 Context 名称和 Endpoint。
- Run 和 Cleanup 使用同一 Context 解析规则；非 Unix Endpoint 明确拒绝。
- JIT Config 只存在于 Docker CLI 子进程环境和容器配置中，不出现在命令参数、结构化日志或持久缓存。

## Socket permissions

OrbStack 的宿主 Socket 在容器中映射为 `root:root 0660`。Run 参数加入 root 附加组，并在可取得宿主 Socket GID 时同时加入该 GID。Runner 继续使用镜像中的非 root `runner` 用户。

## Cache ownership

语言和 Tool Cache 使用 `cache_root` 下的宿主目录 Bind Mount，便于限定共享范围、备份和容量管理。BuildKit 状态属于宿主 Docker Engine，不挂载到 Runner 文件系统，也不由 Backend Cleanup 管理。

## Image choice

不直接使用官方 Runner 镜像作为最终镜像：固定版本官方镜像包含 Runner、Docker CLI 和 Buildx，但缺少 Compose CLI，且包含 Dockerd。项目自有薄层镜像以官方固定版本为基础，添加固定 Compose 并删除 Daemon 可执行文件，避免复制 Runner 安装逻辑。

## Cleanup safety

管理标签是资源所有权的唯一依据。名称用于诊断和 Runner 身份，不作为 CleanupAll 所有权判断。`docker rm -f -v` 只删除容器及匿名卷；Docker 不会因此删除命名卷，Bind Mount 和 BuildKit 状态也不受影响。

## API compatibility

新增 `BACKEND_DOCKER_HOST` Proto 枚举和 Dashboard 映射，避免管理 API 把新 Backend 显示为 unknown。该变更只扩展枚举，不重构 Dashboard。

## Rollback

删除 `docker-host` Runner Set 或改回 `backend: docker` 即可恢复旧行为。新 Backend 不迁移数据库、不修改现有 Backend 资源格式；遗留容器可通过管理标签识别。
