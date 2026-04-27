# Design

## Overview

`cf-flag` is a minimal feature-flag backend implemented in Go and deployed on Cloudflare Workers. The Worker exposes HTTP routes for creating users, creating flags, listing flags, and checking whether a flag is active for a given user.

The system is intentionally simple:

- One Cloudflare Worker handles all API routes.
- One Cloudflare D1 database stores flags.
- Percentage-based flag assignment is computed on demand and is not persisted.

## Goals

- Keep the data model small.
- Support country-based rules, percentage-based rules, and their combination.
- Avoid a `flag_user` mapping table.
- Make percentage assignment deterministic and stable.
- Keep the deployment model simple for Cloudflare Workers.

## Data Model

### `flags`

- `id TEXT PRIMARY KEY`
- `name TEXT NOT NULL`
- `rule TEXT NOT NULL`

The `rule` column is encoded as one of:

- `country:IN`
- `pct:25`
- `country_pct:IN:25`

## API

- `POST /createflag`
- `GET /listflag`
- `POST /checkflag`
- `GET /healthz`

## Rule Evaluation

### Country Rule

For `country:XX`, the Worker compares the `userCountry` provided in the request with the flag country.

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

### Country Percentage Rule

For `country_pct:XX:N`, the Worker first evaluates the country match. If the `userCountry` provided in the request matches `XX`, it then evaluates the percentage assignment (`N`) exactly as described in the Percentage Rule section.

### Percentage Mapping Alternatives

The current implementation is intentionally stateless. It maps a user to a percentage flag by hashing only:

- `flagID`
- `userID`
- `FLAG_HASH_KEY`

This is simple and fast, but it gives an approximately correct percentage, not an exact percentage of the current user population.

Other options:

#### Deterministic Ranking Across All Users

For each user, compute a deterministic score from `flagID` and `userID`, sort all users by score, and activate the top `ceil(N% * total_users)`.

Pros:

- exact percentage over the current user set
- deterministic selection

Cons:

- requires access to the full user population
- expensive to compute on demand
- adding users can shift the cutoff and change who is active

#### Stored Explicit Assignments

At percentage-flag creation time, compute the selected users once and store explicit `flag -> user` assignments.

Pros:

- exact percentage
- fast status checks
- stable assignments over time

Cons:

- requires a mapping table
- adds write-time complexity
- needs a policy for newly created users

### Design Choice

This service uses stateless HMAC bucketing because the goal is to keep the backend minimal and avoid a `flag_user` mapping table.

If future requirements need exact percentage rollouts instead of approximate rollouts, the preferred upgrade path is stored explicit assignments.

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
