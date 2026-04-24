# cf-flag

Minimal feature-flag backend in Go on Cloudflare Workers.

## What it supports

- Create user (`id`, `name`, `email`, `country`).
- Create flag (`id`, `name`, rule as country or percentage).
- Check if flag is active for a user.

Rules are stored as:

- `country:IN`
- `pct:25`

For percentage rules, activation is deterministic using `HMAC-SHA256(FLAG_HASH_KEY, flagID + ":" + userID)`.

## Endpoints

- `POST /users`
- `POST /flags`
- `GET /flags/{flagID}/users/{userID}/active`
- `GET /healthz`

### Create user

```json
POST /users
{
  "id": "user-1",
  "name": "Vaibhav",
  "email": "vaibhav@example.com",
  "country": "IN"
}
```

### Create country flag

```json
POST /flags
{
  "id": "flag-country-in",
  "name": "India users",
  "country": "IN"
}
```

### Create percentage flag

```json
POST /flags
{
  "id": "flag-rollout-25",
  "name": "Rollout 25",
  "percentage": 25
}
```

### Check active

```text
GET /flags/flag-rollout-25/users/user-1/active
```

Response:

```json
{
  "flagId": "flag-rollout-25",
  "userId": "user-1",
  "rule": "pct:25",
  "active": true
}
```

## Setup

1. Install dependencies:

```bash
npm install
go mod tidy
```

2. Create D1 DB:

```bash
wrangler d1 create cf-flag
```

3. Put returned `database_id` into `wrangler.jsonc`.

4. Init local DB schema:

```bash
npm run db:init
```

5. Run locally:

```bash
npm start
```

6. Deploy:

```bash
npm run deploy
```

## Secret key

Do not keep production hash key in `vars`.

Set secret instead:

```bash
wrangler secret put FLAG_HASH_KEY
```

Changing this secret reshuffles percentage cohorts.
