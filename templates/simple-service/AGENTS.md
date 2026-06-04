# simple-service template rules

- This template is for lightweight single-purpose frontend services: landing pages, query tools, converters, generators, calculators, and similar no-backend apps.
- Use React + Vite + TypeScript for app code.
- Use Tailwind CSS v4 utilities for styling.
- Do not add a backend server. Production runtime is static files served by unprivileged Nginx.
- When complex accessible interactions are needed, install `@base-ui/react` and read `https://base-ui.com/llms.txt` before writing components.
- Base UI is not Radix UI: do not use Radix-only APIs like `asChild`; Base UI uses the `render` prop for composition.
- Keep each copied app focused on one purpose. Do not turn a simple-service app into a multi-tool platform.

## Creating a new service

From the `containers` repo root:

```bash
cp -r templates/simple-service apps/<service-name>
```

Then update:

- `apps/<service-name>/docker-bake.hcl`: set `APP` to `<service-name>`.
- `apps/<service-name>/package.json`: set `name` to `<service-name>`.
- `apps/<service-name>/index.html`: set `<title>` and meta description.
- `apps/<service-name>/src/App.tsx`: replace starter content with the service UI.

After the image exists, add the Kubernetes deployment under:

```text
k8s/apps/common/simple-service/<service-name>/
```

and register it in:

```text
k8s/apps/common/simple-service/ks.yaml
```
