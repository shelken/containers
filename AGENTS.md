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
- Dockerfile 里写死的镜像引用（`FROM`、`# syntax=`）写到 patch 版本并钉 digest（`镜像:1.2.3-alpine@sha256:...`），禁止 `latest` 与 `1-alpine` 这类浮动 tag：浮动 tag 的补丁更新不会生成写明版本号的 Renovate PR，构建出的镜像也看不出基底版本；经 bake 变量传入的版本同样写到 patch
