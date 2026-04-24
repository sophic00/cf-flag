# cf-flag

Minimal feature-flag backend in Go on Cloudflare Workers.

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

- `POST /users`
- `POST /flags`
- `GET /flags`
- `GET /flags/{flagID}/users/{userID}/active`
- `GET /healthz`

### Create user

```json
POST /users
{
  "name": "Vaibhav",
  "email": "vaibhav@example.com",
  "country": "IN"
}
```

### Create country flag

```json
POST /flags
{
  "name": "India users",
  "country": "IN"
}
```

### Create percentage flag

```json
POST /flags
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
GET /flags
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

2. Create D1 DB:

```bash
wrangler d1 create cf-flag
```

3. Put returned `database_id` into `wrangler.jsonc`.

4. Init local DB schema:

```bash
make db-init
```

5. Run locally:

```bash
make dev
```

6. Deploy:

```bash
make deploy
```

## Secret key

Do not keep production hash key in `vars`.

Set secret instead:

```bash
wrangler secret put FLAG_HASH_KEY
```

Changing this secret reshuffles percentage cohorts.
