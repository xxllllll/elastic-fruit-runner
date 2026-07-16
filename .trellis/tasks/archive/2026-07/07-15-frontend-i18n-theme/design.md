# 技术设计：前端国际化与亮色主题

## 1. 边界

- 修改范围限定在 `dashboard/src/` 与 `dashboard/index.html`，必要时更新 `dashboard/README.md` 的功能说明。
- 不修改后端接口、数据类型、轮询逻辑和 `doc-site/`。
- 当前仓库没有前端专用 Trellis spec；遵循现有 React/TypeScript 结构与 `npm run build`、`npm run lint` 约定。

## 2. 国际化设计

新增集中式翻译模块，例如 `dashboard/src/i18n.ts`：

- 定义 `Locale = 'en' | 'zh-CN'` 和完整的英文、简体中文字典。
- 使用点号键和简单插值函数处理动态值，例如范围、数量、分钟数和时间后缀。
- 将语言状态、初始化和持久化封装在 UI 偏好模块中，首选 `localStorage`，没有用户选择时读取 `navigator.language`。
- 语言切换时执行 `document.documentElement.lang = locale`。
- 页面与子组件只读取翻译函数，不再新增裸露界面文案。

## 3. 主题设计

新增 UI 偏好状态模块，例如 `dashboard/src/store/useUiPreferences.ts`：

- 定义 `Theme = 'dark' | 'light'`。
- 首次加载读取 `localStorage`；没有值时使用 `window.matchMedia('(prefers-color-scheme: light)')`。
- 切换主题时更新 `document.documentElement.dataset.theme`，并持久化用户选择。
- 监听系统主题变化只影响尚未手动选择主题的用户，避免覆盖用户明确选择。

在 `dashboard/src/index.css` 中建立完整的主题变量：

- 背景、面板、边框、主次文字、强调色、危险色和警告色。
- 进度条、滚动条、分隔线、状态标签和加载动画使用变量。
- `[data-theme='light']` 覆盖变量，保持原有深色变量作为默认值。
- 将组件中的硬编码颜色替换为 `var(--...)` 或语义状态变量。

`PixelPet` 的 Canvas 不能直接读取 CSS 变量作为 `fillStyle`，因此在绘制前根据当前主题选择对应的调色板，并在主题变化时重绘当前帧。

## 4. UI 入口

在现有 Header 右侧增加两个紧凑控制：

- 语言：显示当前语言并允许切换 English / 简体中文。
- 主题：显示当前主题并切换深色 / 亮色。

控件使用原生可访问元素（`button` 或 `select`），提供 `aria-label`，不改变现有网格布局；在移动宽度下允许自然换行或压缩间距。

## 5. 兼容性与风险

- `localStorage` 和 `matchMedia` 仅在浏览器运行时访问，避免 Vite 构建阶段触发 SSR 类问题。
- 翻译键缺失时回退到英文，避免新增状态下出现空白。
- 中文文本可能比英文更长，需检查 Header、状态标签、表头和移动端布局的溢出。
- 主题变量必须覆盖所有高频硬编码颜色，否则亮色主题会出现局部深色残留。

## 6. 回滚

改动集中在前端资源；回滚时恢复 `dashboard/src`、`dashboard/index.html` 和文档变更即可，不涉及数据迁移或后端兼容层。
