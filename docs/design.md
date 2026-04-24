# Design

## Overview

`cf-flag` is a minimal feature-flag backend implemented in Go and deployed on Cloudflare Workers. The Worker exposes HTTP routes for creating users, creating flags, listing flags, and checking whether a flag is active for a given user.

The system is intentionally simple:

- One Cloudflare Worker handles all API routes.
- One Cloudflare D1 database stores users and flags.
- Percentage-based flag assignment is computed on demand and is not persisted.

## Goals

- Keep the data model small.
- Support country-based rules and percentage-based rules.
- Avoid a `flag_user` mapping table.
- Make percentage assignment deterministic and stable.
- Keep the deployment model simple for Cloudflare Workers.

## Data Model

### `users`

- `id TEXT PRIMARY KEY`
- `name TEXT NOT NULL`
- `email TEXT NOT NULL UNIQUE`
- `country TEXT NOT NULL`

### `flags`

- `id TEXT PRIMARY KEY`
- `name TEXT NOT NULL`
- `rule TEXT NOT NULL`

The `rule` column is encoded as one of:

- `country:IN`
- `pct:25`

## API

- `POST /createuser`
- `POST /createflag`
- `GET /listflag`
- `GET /flags/{flagID}/users/{userID}/active`
- `GET /healthz`

## Rule Evaluation

### Country Rule

For `country:XX`, the Worker loads the user by `userID` and compares `users.country` with the flag country.

### Percentage Rule

For `pct:N`, the Worker computes the result from:

- `flagID`
- `userID`
- `FLAG_HASH_KEY`

Algorithm:

1. Build HMAC-SHA256 over `flagID + ":" + userID`
2. Read the first 8 bytes as an unsigned integer
3. Compute `bucket = hash % 10000`
4. Mark active when `bucket < N * 100`

This provides:

- deterministic results for the same user and flag
- stable rollouts over time
- no stored assignment table
- approximately uniform spread for percentage flags

## ID Strategy

User IDs and flag IDs are generated on the backend using UUIDv7 and prefixed for readability:

- `usr_<uuidv7>`
- `flg_<uuidv7>`

UUIDv7 is used because it is stable, unique, and behaves well as input to the deterministic percentage hashing logic.

## Worker Runtime

The Worker uses:

- `github.com/syumai/workers`
- `github.com/syumai/workers/cloudflare/d1`

`main.go` is compiled only for `js/wasm` and serves the `net/http` mux inside the Worker runtime. `main_nonwasm.go` exists only so local `go test` works without importing Worker-only dependencies on native builds.

## Storage and Local Development

- Local development uses Wrangler/Miniflare local D1 state.
- Production uses a live D1 database bound as `DB` in `wrangler.jsonc`.

Local schema init:

```bash
make db-init
```

Live database creation:

```bash
make db-create
```

After creation, copy the returned `database_id` into `wrangler.jsonc`, then initialize the remote schema:

```bash
make db-init-remote
```

To clear data while keeping the schema:

```bash
make db-clean
make db-clean-remote
```

## Deployment Flow

1. `npm install`
2. `go mod tidy`
3. `make db-create`
4. Update `wrangler.jsonc` with the new `database_id`
5. `make db-init-remote`
6. `npx wrangler secret put FLAG_HASH_KEY`
7. `make deploy`

## Tradeoffs

- The active-check route currently works by `flagID` and `userID`, not by `flag name`.
- There is no per-user assignment persistence, so percentage rollouts are approximate rather than exact-count based.
- Email validation is practical and regex-based, not full RFC validation.

These tradeoffs keep the system small and suitable for the current requirements.
