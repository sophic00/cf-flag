#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8787}"
TOTAL_USERS="${TOTAL_USERS:-200}"
PERCENTAGE="${PERCENTAGE:-25}"
COUNTRY_CODE="${COUNTRY_CODE:-IN}"
OTHER_COUNTRY="${OTHER_COUNTRY:-US}"

require_bin() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

get_json() {
  local path="$1"
  curl -fsS "${BASE_URL}${path}"
}

post_json() {
  local path="$1"
  local payload="$2"
  curl -fsS \
    -X POST \
    -H 'content-type: application/json' \
    -d "$payload" \
    "${BASE_URL}${path}"
}

assert_eq() {
  local expected="$1"
  local actual="$2"
  local message="$3"
  if [[ "$expected" != "$actual" ]]; then
    printf 'assertion failed: %s (expected=%s actual=%s)\n' "$message" "$expected" "$actual" >&2
    exit 1
  fi
}

require_bin curl
require_bin jq

if ! [[ "$TOTAL_USERS" =~ ^[0-9]+$ ]] || (( TOTAL_USERS < 100 )); then
  printf 'TOTAL_USERS must be an integer >= 100\n' >&2
  exit 1
fi

if ! [[ "$PERCENTAGE" =~ ^[0-9]+$ ]] || (( PERCENTAGE < 0 || PERCENTAGE > 100 )); then
  printf 'PERCENTAGE must be an integer between 0 and 100\n' >&2
  exit 1
fi

printf 'Checking health at %s/healthz\n' "$BASE_URL"
health="$(get_json '/healthz')"
assert_eq 'true' "$(jq -r '.ok' <<<"$health")" 'health check'

run_id="$(date +%s)"
split_index=$(( TOTAL_USERS / 2 ))
expected_country_matches="$split_index"

declare -a user_ids
declare -a user_countries

printf 'Generating %d users locally\n' "$TOTAL_USERS"
for ((i = 1; i <= TOTAL_USERS; i++)); do
  if (( i <= split_index )); then
    country="$COUNTRY_CODE"
  else
    country="$OTHER_COUNTRY"
  fi

  user_ids+=("usr_test_${run_id}_${i}")
  user_countries+=("$country")
done

country_flag_name="country-${COUNTRY_CODE}-${run_id}"
pct_flag_name="rollout-${PERCENTAGE}-${run_id}"

printf 'Creating country flag %s\n' "$country_flag_name"
country_flag_payload="$(jq -nc \
  --arg name "$country_flag_name" \
  --arg country "$COUNTRY_CODE" \
  '{name:$name, country:$country}')"
country_flag_response="$(post_json '/createflag' "$country_flag_payload")"
country_flag_id="$(jq -r '.flag.id' <<<"$country_flag_response")"

printf 'Creating percentage flag %s\n' "$pct_flag_name"
pct_flag_payload="$(jq -nc \
  --arg name "$pct_flag_name" \
  --argjson percentage "$PERCENTAGE" \
  '{name:$name, percentage:$percentage}')"
pct_flag_response="$(post_json '/createflag' "$pct_flag_payload")"
pct_flag_id="$(jq -r '.flag.id' <<<"$pct_flag_response")"

country_pct_flag_name="country-pct-${COUNTRY_CODE}-${PERCENTAGE}-${run_id}"

printf 'Creating country+percentage flag %s\n' "$country_pct_flag_name"
country_pct_flag_payload="$(jq -nc \
  --arg name "$country_pct_flag_name" \
  --arg country "$COUNTRY_CODE" \
  --argjson percentage "$PERCENTAGE" \
  '{name:$name, country:$country, percentage:$percentage}')"
country_pct_flag_response="$(post_json '/createflag' "$country_pct_flag_payload")"
country_pct_flag_id="$(jq -r '.flag.id' <<<"$country_pct_flag_response")"

printf 'Verifying /listflag contains created flags\n'
flags_response="$(get_json '/listflag')"
country_found="$(jq -r --arg id "$country_flag_id" '[.flags[] | select(.id == $id)] | length' <<<"$flags_response")"
pct_found="$(jq -r --arg id "$pct_flag_id" '[.flags[] | select(.id == $id)] | length' <<<"$flags_response")"
country_pct_found="$(jq -r --arg id "$country_pct_flag_id" '[.flags[] | select(.id == $id)] | length' <<<"$flags_response")"
assert_eq '1' "$country_found" 'country flag listed'
assert_eq '1' "$pct_found" 'percentage flag listed'
assert_eq '1' "$country_pct_found" 'country+percentage flag listed'

printf 'Checking country rule across %d users\n' "$TOTAL_USERS"
country_active_count=0
for ((idx = 0; idx < TOTAL_USERS; idx++)); do
  user_id="${user_ids[$idx]}"
  country="${user_countries[$idx]}"
  payload="$(jq -nc \
    --arg flagId "$country_flag_id" \
    --arg userId "$user_id" \
    --arg userCountry "$country" \
    '{flagId:$flagId, userId:$userId, userCountry:$userCountry}')"
  response="$(post_json '/checkflag' "$payload")"
  active="$(jq -r '.active' <<<"$response")"

  expected='false'
  if [[ "$country" == "$COUNTRY_CODE" ]]; then
    expected='true'
    ((country_active_count += 1))
  fi

  assert_eq "$expected" "$active" "country flag for ${user_id}"

  if (((idx + 1) % 25 == 0)); then
    printf '  checked %d/%d country evaluations\n' "$((idx + 1))" "$TOTAL_USERS"
  fi
done
assert_eq "$expected_country_matches" "$country_active_count" 'country active count'

printf 'Checking percentage rule across %d users\n' "$TOTAL_USERS"
pct_active_count=0
for ((idx = 0; idx < TOTAL_USERS; idx++)); do
  user_id="${user_ids[$idx]}"
  country="${user_countries[$idx]}"
  payload="$(jq -nc \
    --arg flagId "$pct_flag_id" \
    --arg userId "$user_id" \
    --arg userCountry "$country" \
    '{flagId:$flagId, userId:$userId, userCountry:$userCountry}')"
  response="$(post_json '/checkflag' "$payload")"
  active="$(jq -r '.active' <<<"$response")"
  if [[ "$active" == 'true' ]]; then
    ((pct_active_count += 1))
  fi

  if (( idx < 10 )); then
    repeat_response="$(post_json '/checkflag' "$payload")"
    repeat_active="$(jq -r '.active' <<<"$repeat_response")"
    assert_eq "$active" "$repeat_active" "deterministic percentage flag for ${user_id}"
  fi

  if (((idx + 1) % 25 == 0)); then
    printf '  checked %d/%d percentage evaluations\n' "$((idx + 1))" "$TOTAL_USERS"
  fi
done

expected_pct_count=$(( TOTAL_USERS * PERCENTAGE / 100 ))
tolerance=$(( TOTAL_USERS / 12 ))
if (( tolerance < 10 )); then
  tolerance=10
fi
min_pct_count=$(( expected_pct_count - tolerance ))
max_pct_count=$(( expected_pct_count + tolerance ))
if (( min_pct_count < 0 )); then
  min_pct_count=0
fi
if (( max_pct_count > TOTAL_USERS )); then
  max_pct_count=$TOTAL_USERS
fi

if (( pct_active_count < min_pct_count || pct_active_count > max_pct_count )); then
  printf 'percentage rollout out of expected range: active=%d expected=%d range=[%d,%d]\n' \
    "$pct_active_count" "$expected_pct_count" "$min_pct_count" "$max_pct_count" >&2
  exit 1
fi

printf 'Checking country+percentage rule across %d users\n' "$TOTAL_USERS"
country_pct_active_count=0
for ((idx = 0; idx < TOTAL_USERS; idx++)); do
  user_id="${user_ids[$idx]}"
  country="${user_countries[$idx]}"
  payload="$(jq -nc \
    --arg flagId "$country_pct_flag_id" \
    --arg userId "$user_id" \
    --arg userCountry "$country" \
    '{flagId:$flagId, userId:$userId, userCountry:$userCountry}')"
  response="$(post_json '/checkflag' "$payload")"
  active="$(jq -r '.active' <<<"$response")"

  if [[ "$country" != "$COUNTRY_CODE" ]]; then
    assert_eq "false" "$active" "country+pct flag should be false for non-matching country ${country}"
  elif [[ "$active" == 'true' ]]; then
    ((country_pct_active_count += 1))
  fi

  if (((idx + 1) % 25 == 0)); then
    printf '  checked %d/%d country+percentage evaluations\n' "$((idx + 1))" "$TOTAL_USERS"
  fi
done

expected_country_pct_count=$(( expected_country_matches * PERCENTAGE / 100 ))
min_country_pct_count=$(( expected_country_pct_count - tolerance ))
max_country_pct_count=$(( expected_country_pct_count + tolerance ))
if (( min_country_pct_count < 0 )); then
  min_country_pct_count=0
fi
if (( max_country_pct_count > expected_country_matches )); then
  max_country_pct_count=$expected_country_matches
fi

if (( country_pct_active_count < min_country_pct_count || country_pct_active_count > max_country_pct_count )); then
  printf 'country+percentage rollout out of expected range: active=%d expected=%d range=[%d,%d]\n' \
    "$country_pct_active_count" "$expected_country_pct_count" "$min_country_pct_count" "$max_country_pct_count" >&2
  exit 1
fi

printf '\nSmoke test passed\n'
printf '  users created: %d\n' "$TOTAL_USERS"
printf '  country flag id: %s\n' "$country_flag_id"
printf '  percentage flag id: %s\n' "$pct_flag_id"
printf '  country matches: %d\n' "$country_active_count"
printf '  percentage actives: %d (expected %d, range [%d,%d])\n' \
  "$pct_active_count" "$expected_pct_count" "$min_pct_count" "$max_pct_count"
printf '  country+percentage flag id: %s\n' "$country_pct_flag_id"
printf '  country+percentage actives: %d (expected %d, range [%d,%d])\n' \
  "$country_pct_active_count" "$expected_country_pct_count" "$min_country_pct_count" "$max_country_pct_count"
