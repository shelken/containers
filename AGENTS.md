## 项目原则

- 使用 `Conventional Commits` 规范：`<type>[optional scope]: <description>`，提交信息应分点简洁，**标题英文，内容中文**

## 做什么

- 主要用于构建镜像/容器化服务
- 通过统一的CI/CD流程打包服务镜像 `.github/workflows/release.yml`
- `apps` 下包含了简单的容器化操作例如`apps/caddy`,用一个Dockerfile构建外部任意服务; 也有形如`apps/zte-mifi-exporter`这种简单的自维护服务

## 不做什么

- 过于复杂的服务不该在这个仓库

## 版本约定

- 基础镜像版本来自 `docker-bake.hcl` 的变量，注解用 `datasource=docker`，值与 Docker 标签一致（不带 `v`）
- Dockerfile 只通过 `ARG` 接收 bake 传入的版本，不依赖 `DOCKER_META_VERSION`：PR 构建时它是 `sandbox`
