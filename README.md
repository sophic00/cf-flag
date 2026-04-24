# cf-flag

Minimal feature-flag backend in Go on Cloudflare Workers.

Design details are documented in `docs/design.md`.

## What it supports

- Create user (`name`, `email`, `country`) with backend-generated `id`.
- Create flag (`name`, rule as country or percentage) with backend-generated `id`.
- Check if flag is active for a user.

Rules are stored as:

- `country:IN`
- `pct:25`

For percentage rules, activation is deterministic using `HMAC-SHA256(FLAG_HASH_KEY, flagID + ":" + userID)`.

User and flag IDs are generated as UUIDv7 (`usr_<uuidv7>`, `flg_<uuidv7>`), which are stable and distribute well for percentage hashing.

## Endpoints

- `POST /createuser`
- `POST /createflag`
- `GET /listflag`
- `GET /flags/{flagID}/users/{userID}/active`
- `GET /healthz`

### Create user

```json
POST /createuser
{
  "name": "Vaibhav",
  "email": "vaibhav@example.com",
  "country": "IN"
}
```

### Create country flag

```json
POST /createflag
{
  "name": "India users",
  "country": "IN"
}
```

### Create percentage flag

```json
POST /createflag
{
  "name": "Rollout 25",
  "percentage": 25
}
```

### Check active

```text
GET /flags/flg_<generated-id>/users/usr_<generated-id>/active
```

Response:

```json
{
  "flagId": "flg_01969587-83da-72a6-b8ef-f6f8ef986355",
  "userId": "usr_01969587-8428-7738-9ec1-cd0df1278d5e",
  "rule": "pct:25",
  "active": true
}
```

### List all flags

```text
GET /listflag
```

Response:

```json
{
  "flags": [
    {
      "id": "flg_01969587-83da-72a6-b8ef-f6f8ef986355",
      "name": "Rollout 25",
      "rule": "pct:25"
    },
    {
      "id": "flg_01969588-a71c-7fcb-b2fe-cf6028dc1f4e",
      "name": "India users",
      "rule": "country:IN"
    }
  ]
}
```

## Setup

1. Install dependencies:

```bash
npm install
go mod tidy
```

Or using Make:

```bash
make test
```

Smoke test against a running local or deployed Worker:

```bash
BASE_URL=http://127.0.0.1:8787 TOTAL_USERS=200 make smoke-test
```

The script creates a few hundred users, creates one country flag and one percentage flag, verifies `/listflag`, checks all country matches, and verifies the percentage rollout stays within a reasonable range.

2. Create D1 DB:

```bash
make db-create
```

This creates the live Cloudflare D1 database. Wrangler prints the new `database_id`; copy that value into `wrangler.jsonc` under `d1_databases[0].database_id`.

If you want a different database name:

```bash
DB_NAME=my-live-db make db-create
```

3. Put returned `database_id` into `wrangler.jsonc`.

4. Init local DB schema:

```bash
make db-init
```

To wipe all data from the local DB while keeping the schema:

```bash
make db-clean
```

5. Run locally:

```bash
make dev
```

6. Deploy:

```bash
make deploy
```

For the remote database schema, run:

```bash
make db-init-remote
```

To wipe all data from the live database while keeping the schema:

```bash
make db-clean-remote
```

## Secret key

Do not keep production hash key in `vars`.

Set secret instead:

```bash
wrangler secret put FLAG_HASH_KEY
```

Changing this secret reshuffles percentage cohorts.

## GitHub Actions Deploy

`.github/workflows/deploy-prod.yml` deploys automatically when code is pushed to the `prod` branch.

The workflow deploys to the GitHub `production` environment, so deployments show up in GitHub's Deployments/Environments UI.

Set these GitHub repository secrets before relying on the workflow:

- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ACCOUNT_ID`

Set this GitHub repository variable if you want the environment to show the live Worker URL:

- `PRODUCTION_URL`

The Worker secret `FLAG_HASH_KEY` is not managed by GitHub Actions. Set it once in Cloudflare with:

```bash
npx wrangler secret put FLAG_HASH_KEY
```
