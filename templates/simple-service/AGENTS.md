# simple-service template rules

- This template is for lightweight single-purpose frontend services: landing pages, query tools, converters, generators, calculators, and similar no-backend apps.
- Use React + Vite + TypeScript for app code.
- Use Tailwind CSS v4 utilities for styling.
- Do not add a backend server. Production runtime is static files served by unprivileged Nginx.
- When complex accessible interactions are needed, install `@base-ui/react` and read `https://base-ui.com/llms.txt` before writing components.
- Base UI is not Radix UI: do not use Radix-only APIs like `asChild`; Base UI uses the `render` prop for composition.
- Keep each copied app focused on one purpose. Do not turn a simple-service app into a multi-tool platform.
