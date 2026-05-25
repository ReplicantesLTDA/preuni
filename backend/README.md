# preuni backend

Single Go monolith. Two modules in the workspace:

- `pkg/` — shared infrastructure (logger, errors, middleware, config)
- `app/` — the deployable binary; domains under `app/internal/<domain>/`

Run, test, and deploy from `app/`. See [`app/README.md`](./app/README.md) for details.
