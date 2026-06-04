## 项目原则

- 使用 `Conventional Commits` 规范：`<type>[optional scope]: <description>`，提交信息应分点简洁，**标题英文，内容中文**
- 新增或更新 npm 依赖时，必须尊重本机 npm minimum release age / `before` 配置；选择配置允许范围内的最新稳定 semver 版本，不绕过 release-age 直接使用刚发布版本。
