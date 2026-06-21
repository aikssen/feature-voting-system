#!/usr/bin/env bash
#
# End-to-end HTTP smoke test for the SoundFlow backend (TASKS.md Phase 4).
#
# Exercises the full contract against a running API + Postgres:
#   signup -> login -> discovery (sorts/search) -> create -> duplicate guard
#   -> vote -> self-vote/duplicate-vote/auth rejections -> ranking -> CORS.
#
# Usage:
#   API_BASE=http://localhost:3000/api/v1 ./scripts/smoke.sh
#
# Requires: curl, python3 (for JSON parsing). Exits non-zero on first failure.
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:3000/api/v1}"
ORIGIN="${ORIGIN:-http://localhost:5173}"
PASS=0
FAIL=0

# Unique suffix so the script is idempotent across runs (emails must be unique).
SUFFIX="$(date +%s)$RANDOM"

green() { printf '\033[32m%s\033[0m\n' "$1"; }
red() { printf '\033[31m%s\033[0m\n' "$1"; }

# jget <json> <python-expression-on-`d`>
jget() { python3 -c "import sys,json; d=json.load(sys.stdin); print($2)" <<<"$1"; }

# check <description> <actual> <expected>
check() {
  if [ "$2" = "$3" ]; then
    green "  PASS  $1 ($2)"
    PASS=$((PASS + 1))
  else
    red "  FAIL  $1 — expected '$3', got '$2'"
    FAIL=$((FAIL + 1))
  fi
}

# status <METHOD> <path> [json-body] [bearer-token]
status() {
  local method="$1" path="$2" body="${3:-}" token="${4:-}"
  local args=(-s -o /dev/null -w '%{http_code}' -X "$method" "$API_BASE$path")
  [ -n "$body" ] && args+=(-H 'Content-Type: application/json' -d "$body")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  curl "${args[@]}"
}

# body <METHOD> <path> [json-body] [bearer-token]
body() {
  local method="$1" path="$2" body="${3:-}" token="${4:-}"
  local args=(-s -X "$method" "$API_BASE$path")
  [ -n "$body" ] && args+=(-H 'Content-Type: application/json' -d "$body")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  curl "${args[@]}"
}

echo "== SoundFlow E2E smoke test =="
echo "API: $API_BASE"
echo

echo "[1] Health"
check "health 200" "$(status GET /health)" "200"

echo "[2] Auth — signup (auto-login) + login"
EMAIL_A="alice+$SUFFIX@example.com"
EMAIL_B="bob+$SUFFIX@example.com"
SIGNUP_A=$(body POST /auth/signup "{\"name\":\"Alice\",\"email\":\"$EMAIL_A\",\"password\":\"Pa\$s\"}")
TOKEN_A=$(jget "$SIGNUP_A" "d['token']")
check "signup returns a token" "$([ -n "$TOKEN_A" ] && echo yes || echo no)" "yes"
check "signup returns the user" "$(jget "$SIGNUP_A" "d['user']['email']")" "$EMAIL_A"

# Second user, to vote on Alice's request.
SIGNUP_B=$(body POST /auth/signup "{\"name\":\"Bob\",\"email\":\"$EMAIL_B\",\"password\":\"Pa\$s\"}")
TOKEN_B=$(jget "$SIGNUP_B" "d['token']")

check "weak password rejected (400)" "$(status POST /auth/signup "{\"name\":\"X\",\"email\":\"x+$SUFFIX@e.com\",\"password\":\"weak\"}")" "400"
check "duplicate email rejected (400)" "$(status POST /auth/signup "{\"name\":\"Alice\",\"email\":\"$EMAIL_A\",\"password\":\"Pa\$s\"}")" "400"

LOGIN_A=$(body POST /auth/login "{\"email\":\"$EMAIL_A\",\"password\":\"Pa\$s\"}")
check "login 200" "$([ -n "$(jget "$LOGIN_A" "d['token']")" ] && echo 200 || echo 0)" "200"
check "wrong password rejected (401)" "$(status POST /auth/login "{\"email\":\"$EMAIL_A\",\"password\":\"nope!\"}")" "401"

echo "[3] Discovery — sorts, search, pagination metadata"
LIST=$(body GET "/features?sort=trending")
check "list has pagination wrapper" "$(jget "$LIST" "'items' in d and 'total_pages' in d")" "True"
check "seeded features present" "$(jget "$LIST" "len(d['items'])>=6")" "True"
check "default sort is trending (Crossfade newest+voted on top)" "$(jget "$LIST" "d['items'][0]['title']")" "Crossfade"
MOST=$(body GET "/features?sort=most_voted")
check "most_voted: top has the highest votes" "$(jget "$MOST" "d['items'][0]['total_votes']==max(i['total_votes'] for i in d['items'])")" "True"
NEW=$(body GET "/features?sort=newest")
check "newest: items sorted by created_at desc" "$(jget "$NEW" "all(d['items'][i]['created_at']>=d['items'][i+1]['created_at'] for i in range(len(d['items'])-1))")" "True"
SEARCH=$(body GET "/features?search=offline")
check "search matches title (case-insensitive)" "$(jget "$SEARCH" "any('Offline' in i['title'] for i in d['items'])")" "True"
check "ranking present and 1-based" "$(jget "$LIST" "d['items'][0]['rank']")" "1"

echo "[4] Create feature (auth required) + duplicate guard"
check "create unauthenticated rejected (401)" "$(status POST /features '{"title":"No auth","description":"should fail"}')" "401"
TITLE="Smart Shuffle $SUFFIX"
CREATE=$(body POST /features "{\"title\":\"$TITLE\",\"description\":\"Shuffle that learns your taste.\"}" "$TOKEN_A")
FID=$(jget "$CREATE" "d['id']")
check "create 201 returns id" "$([ -n "$FID" ] && echo yes || echo no)" "yes"
check "creator flagged is_author" "$(jget "$CREATE" "d['is_author']")" "True"
check "duplicate title rejected (409)" "$(status POST /features "{\"title\":\"$TITLE\",\"description\":\"dupe\"}" "$TOKEN_A")" "409"
check "short title rejected (400)" "$(status POST /features '{"title":"x","description":"too short title"}' "$TOKEN_A")" "400"

echo "[5] Voting + business rules"
check "vote unauthenticated rejected (401)" "$(status POST "/features/$FID/vote")" "401"
check "self-vote rejected (403)" "$(status POST "/features/$FID/vote" "" "$TOKEN_A")" "403"
VOTE=$(body POST "/features/$FID/vote" "" "$TOKEN_B")
check "Bob votes on Alice's feature (count=1)" "$(jget "$VOTE" "d['total_votes']")" "1"
check "vote sets has_voted" "$(jget "$VOTE" "d['has_voted']")" "True"
check "duplicate vote rejected (409)" "$(status POST "/features/$FID/vote" "" "$TOKEN_B")" "409"
check "vote on missing feature (404)" "$(status POST "/features/00000000-0000-0000-0000-000000000000/vote" "" "$TOKEN_B")" "404"

echo "[6] Per-user view flags"
AS_B=$(body GET "/features/$FID" "" "$TOKEN_B")
check "Bob sees has_voted=true on this feature" "$(jget "$AS_B" "d['has_voted']")" "True"
check "Bob sees is_author=false" "$(jget "$AS_B" "d['is_author']")" "False"
ANON=$(body GET "/features/$FID")
check "anonymous sees has_voted=false" "$(jget "$ANON" "d['has_voted']")" "False"

echo "[7] CORS preflight for the frontend origin"
ACAO=$(curl -s -o /dev/null -D - -X OPTIONS "$API_BASE/features" \
  -H "Origin: $ORIGIN" -H 'Access-Control-Request-Method: POST' \
  -H 'Access-Control-Request-Headers: authorization,content-type' \
  | tr -d '\r' | awk -F': ' 'tolower($1)=="access-control-allow-origin"{print $2}')
check "CORS allows the frontend origin" "$ACAO" "$ORIGIN"

echo
echo "== Results: $PASS passed, $FAIL failed =="
[ "$FAIL" -eq 0 ] || exit 1
