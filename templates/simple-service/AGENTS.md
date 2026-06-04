# simple-service 模板规则

- 此模板用于轻量、单一目的的前端服务：展示页、查询工具、转换器、生成器、计算器等无后端应用。
- 应用代码使用 React + Vite + TypeScript。
- 样式优先使用 Tailwind CSS v4 utilities。
- 不要添加后端服务。生产运行时是 unprivileged Nginx 托管的静态文件。
- 需要 Dialog、Menu、Select、Form、Toast 等复杂可访问性交互时，再安装 `@base-ui/react`，并先阅读 `https://base-ui.com/llms.txt`。
- Base UI 不是 Radix UI：不要使用 Radix 专属 API，例如 `asChild`；Base UI 使用 `render` prop 做组合。
- 每个复制后的应用只服务一个明确目的。不要把 simple-service 应用扩展成多工具平台。
- 复制模板后，必须清掉模板里不再需要的示例内容、说明文字和占位数据；只保留当前服务真正需要的代码、文案和配置。

## 创建新服务

在 `containers` 仓库根目录执行：

```bash
cp -r templates/simple-service apps/<service-name>
```

随后更新：

- `apps/<service-name>/docker-bake.hcl`：将 `APP` 改为 `<service-name>`。
- `apps/<service-name>/package.json`：将 `name` 改为 `<service-name>`。
- `apps/<service-name>/index.html`：更新 `<title>` 和 meta description。
- `apps/<service-name>/src/App.tsx`：删除模板 starter UI，替换为当前服务需要的界面。
- 复制生成的 `apps/<service-name>/AGENTS.md`：删除只对模板创建流程有用、对该服务无长期约束的内容；保留该服务后续开发必须遵守的规则。

镜像可用后，在 home-ops 中添加 Kubernetes 部署目录：

```text
k8s/apps/common/simple-service/<service-name>/
```

并注册到：

```text
k8s/apps/common/simple-service/ks.yaml
```
