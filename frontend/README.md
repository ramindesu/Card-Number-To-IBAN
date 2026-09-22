# شبایار frontend

Nuxt/Vue frontend for the Card-to-IBAN API.

## Run locally

```bash
pnpm install
pnpm dev
```

The Nuxt development server proxies `/api/*` to `http://localhost:8080` by
default. Override it with `NUXT_API_PROXY_TARGET` when the Go API runs on a
different address.

The page calls `POST /api/v1/card-to-iban` and expects the backend response
contract `{ data: ... }` or `{ error: ... }`.
