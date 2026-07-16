# 实施计划：前端国际化与亮色主题

## 顺序清单

1. 新增 UI 偏好状态与初始化逻辑：语言、主题、`localStorage`、系统偏好和 DOM 属性同步。
2. 新增英文 / 简体中文翻译资源及插值工具，整理 App 与组件中的页面文案。
3. 在 Header 增加语言和主题控制，并补充可访问名称、当前状态和切换行为。
4. 把全局 CSS 与组件内硬编码颜色迁移到语义主题变量，完成亮色主题定义。
5. 让 Canvas 像素宠物按主题使用对应调色板，并在主题变化时重绘。
6. 更新 `html lang` 初始值策略和必要的 README 功能说明。
7. 做定向检查：英文 / 中文切换、深色 / 亮色切换、刷新后的持久化、系统偏好初始化、桌面 / 移动布局。

## 预期文件

- `dashboard/src/i18n.ts`
- `dashboard/src/store/useUiPreferences.ts`
- `dashboard/src/App.tsx`
- `dashboard/src/components/ConnectionStatus.tsx`
- `dashboard/src/components/JobRow.tsx`
- `dashboard/src/components/RunnerSetPanel.tsx`
- `dashboard/src/components/SystemVitals.tsx`
- `dashboard/src/components/PixelPet.tsx`
- `dashboard/src/components/petMood.ts`
- `dashboard/src/index.css`
- `dashboard/index.html`
- `dashboard/README.md`（仅在功能说明需要同步时修改）

## 验证命令

```bash
cd dashboard
npm run build
npm run lint
```

必要时再运行仓库级验证：

```bash
make build
```

## 风险检查点

- 翻译覆盖率：不得遗漏加载、错误、表头、空状态、状态标签和动态后缀。
- 主题覆盖率：不得继续使用会破坏亮色主题的主要硬编码颜色。
- 中文布局：检查 Header、状态 badge、统计卡片和 600px 移动端宽度。
- 状态行为：主题 / 语言切换不能触发数据请求或重置 dashboard 数据。
