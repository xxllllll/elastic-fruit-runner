# 前端国际化与亮色主题

## Goal

为内嵌的 React 监控面板增加简体中文界面和明暗主题切换，同时保留现有终端风格、信息密度与响应式布局。

## Background

- 本次范围是 `dashboard/`，该目录由 React 19、TypeScript、Vite 和 Tailwind CSS v4 构建，产物通过 `dashboard/embed.go` 嵌入 Go 服务。
- 当前没有国际化依赖或语言状态；页面文案分散在 `App.tsx`、`components/*.tsx` 和 `components/petMood.ts` 中，`dashboard/index.html` 固定为 `lang="en"`。
- 当前视觉以深色为唯一主题。`index.css` 已定义少量颜色变量，但组件中仍存在大量硬编码颜色，因此仅修改根变量不能形成完整、可读的亮色主题。
- 当前前端基线验证通过：`npm run build` 与 `npm run lint` 均以退出码 0 结束。

## Requirements

1. 增加英文与简体中文两套界面文案，覆盖页面标题、状态、统计项、表头、空状态、加载/错误状态、宠物状态描述和动态时间后缀。
2. 提供明确的语言切换入口，并在切换后同步更新文档语言属性。
3. 增加深色与亮色主题，亮色主题必须覆盖背景、边框、主次文字、进度条、状态色、Canvas 像素宠物等现有视觉元素。
4. 提供明确的主题切换入口；首次访问跟随系统主题偏好，用户手动切换后持久化选择。
5. 保持当前桌面、平板与移动端布局，不改变 API 数据结构和后端接口。
6. 主题和语言控制应集中管理，避免继续在组件内新增不可复用的文案或主题颜色常量。

## Acceptance Criteria

- [x] 用户能在英文和简体中文之间切换，页面主要可见文案不存在遗漏或混合语言。
- [x] 用户能在深色和亮色主题之间切换，所有主要区域在两种主题下均有足够的文字、边框和状态色对比度。
- [x] 语言切换后，`document.documentElement.lang` 与当前语言一致。
- [x] 主题切换不会改变数据请求、轮询、计算结果或页面布局。
- [x] 桌面、平板和移动端仍保持现有响应式结构。
- [x] `npm run build` 与 `npm run lint` 通过。
- [x] 最终运行仓库规定的相关质量检查，并如实记录结果。

## Out of Scope

- 文档站 `doc-site/` 的国际化和主题改造。
- 后端 API、配置文件与数据模型修改。
- 新增简体中文之外的第三种语言。
- 重新设计当前监控面板的信息架构。

## Default Behavior

- 首次访问：浏览器语言为 `zh` / `zh-CN` 时使用简体中文，否则使用英文。
- 首次访问：系统偏好为亮色时使用亮色，否则使用深色。
- 用户手动切换语言或主题后，使用 `localStorage` 记忆选择，并优先于系统偏好。

## Validation Evidence

- `npm run build`：通过，TypeScript 与 Vite 生产构建成功。
- `npm run lint`：通过。
- `make unit-test`：通过，所有 Go 单元测试成功。
- `uvx prek run --all-files`：通过，包括 `go-mod-tidy`、`go fmt`、`go imports` 和 `golangci-lint`。
- 亮色主题在白色面板上的最小主要文本对比度为 5.57:1，危险色与警告色均高于 6:1。
- 最新二进制已部署至 `~/.local/bin/elastic-fruit-runner`，运行版本包含提交 `8cd40ab4902c3c16b037adef3b70d7996f118a6e`。
- 本地服务重启后四个控制面 API 均返回 HTTP 200，嵌入的前端资源包含中英文资源、系统主题检测和持久化键。
