# 修复 Dashboard 历史任务白屏

## Goal

Dashboard 必须能够显示包含 completed-only 历史任务的运行状态，不能因任务
记录缺少 Runner 信息而在初始化渲染阶段白屏。

## Background

- Playwright MCP 在 `http://127.0.0.1:38080/` 稳定复现空页面，
  控制台错误为 `TypeError: Cannot read properties of undefined (reading 'split')`。
- `ListJobRecords` 返回的部分 completed-only 记录缺少 `runnerName` 与
  `runnerSetName`，静态资源及所有初始化 RPC 均返回 `200`。
- 后端在守护进程重启等情况下会通过 `InsertCompletedJob` 写入 Runner 名称为空的记录；Proto3 JSON 会省略默认空字符串字段。
- 前端响应类型错误地把两个字段声明为必填字符串，
  `dashboard/src/components/JobRow.tsx:26` 将缺失值传给
  `dashboard/src/utils.ts:22` 的 `shortName()`，最终调用
  `undefined.split('-')`。

## Requirements

- 将 `ListJobRecords` 的前端响应类型修正为符合运行时数据：Runner 名称字段允许缺失。
- 在 API 数据转换边界把缺失的 Runner 名称规范化为稳定的前端模型值。
- 历史任务没有 Runner 名称时，任务行显示明确的本地化占位文本，不调用 `shortName()`。
- 保持已有完整任务记录的名称缩写、状态和持续时间显示行为不变。
- 不修改 Proto、数据库结构或 completed-only 记录的后端语义。

## Acceptance Criteria

- [x] 包含缺失 `runnerName` / `runnerSetName` 的 `ListJobRecords` 响应时，
  Dashboard 能正常渲染且控制台不再出现 `undefined.split` 异常。
- [x] 缺少 Runner 名称的任务行显示本地化的“未知 Runner”语义文本。
- [x] 完整 Runner 名称仍通过 `shortName()` 按原规则显示。
- [x] Dashboard TypeScript 构建与 ESLint 检查通过。
- [x] 使用 Playwright 对修复后的页面进行运行态验证，页面产生可访问性内容且控制台无错误。

## Validation Evidence

- `pnpm run lint`：通过。
- `pnpm run build`：通过，生成 `dist/assets/index-CIXLUEJU.js`。
- `go vet ./...`：通过。
- `go test -count=1 ./...`：通过。
- `git diff --check`：通过。
- Playwright MCP：通过临时同源代理连接真实 `38080` 后端；页面完整
  渲染，6 条 completed-only 记录在简体中文下显示“未知”、英文下显示
  `UNKNOWN`，验证期间控制台错误为 0。
- `make check`：未完整执行，环境缺少 `golangci-lint`；此前的 Dashboard
  构建、`go vet` 和 Go binary 构建均已通过。
- `make test`：单元测试通过；集成测试因缺少仓库要求的 `.env.integration-test` 未执行。

## Out of Scope

- 清理或回填 SQLite 中已有的 completed-only 记录。
- 改变守护进程重启时记录完成事件的处理方式。
- 引入新的前端测试框架。
