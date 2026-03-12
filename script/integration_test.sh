#!/usr/bin/env bash
set -euo pipefail

# ─── Integration test for the fz CLI ───
# Exercises every fz command against the live Fizzy API.
# Requires: jq, ./fz binary, authenticated session.

# ── Colors ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# ── Prerequisite checks ──
if ! command -v jq &>/dev/null; then
  echo "jq is required but not installed" >&2
  exit 1
fi

if [ ! -x ./fz ]; then
  echo "fz binary not found; run 'make build' first" >&2
  exit 1
fi

if ! ./fz auth status --check &>/dev/null; then
  echo "Not authenticated; run './fz auth login' first" >&2
  exit 1
fi

# ── Globals ──
PASS=0
FAIL=0
TOTAL=0
SUFFIX="fz-test-$$"

# Resource IDs (populated during tests).
BOARD_ID=""
COL1_ID=""
COL2_ID=""
CARD1_NUM=""
CARD2_NUM=""
COMMENT_ID=""
STEP_ID=""
WEBHOOK_ID=""
REACTION_ID=""
COMMENT_REACTION_ID=""
USER_ID=""

# ── Helper: run_test ──
# Usage: run_test "name" "command" ["expected_substring"]
run_test() {
  local name="$1"
  local cmd="$2"
  local expected="${3:-}"

  TOTAL=$((TOTAL + 1))

  local output
  local exit_code=0
  output=$(eval "$cmd" 2>&1) || exit_code=$?

  if [ "$exit_code" -ne 0 ]; then
    FAIL=$((FAIL + 1))
    printf "  ${RED}FAIL${NC} %s (exit code %d)\n" "$name" "$exit_code"
    printf "       output: %s\n" "$output" | head -5
    return
  fi

  if [ -n "$expected" ]; then
    if ! echo "$output" | grep -qF "$expected"; then
      FAIL=$((FAIL + 1))
      printf "  ${RED}FAIL${NC} %s (missing: %s)\n" "$name" "$expected"
      printf "       output: %s\n" "$output" | head -5
      return
    fi
  fi

  PASS=$((PASS + 1))
  printf "  ${GREEN}PASS${NC} %s\n" "$name"
}

# ── Cleanup trap ──
cleanup() {
  echo ""
  echo "=== Cleanup ==="

  if [ -n "$WEBHOOK_ID" ] && [ -n "$BOARD_ID" ]; then
    ./fz webhook delete "$BOARD_ID" "$WEBHOOK_ID" --yes 2>/dev/null || true
  fi

  if [ -n "$COMMENT_ID" ] && [ -n "$CARD1_NUM" ]; then
    ./fz comment delete "$CARD1_NUM" "$COMMENT_ID" --yes 2>/dev/null || true
  fi

  if [ -n "$CARD1_NUM" ]; then
    ./fz card delete "$CARD1_NUM" --yes 2>/dev/null || true
  fi

  if [ -n "$CARD2_NUM" ]; then
    ./fz card delete "$CARD2_NUM" --yes 2>/dev/null || true
  fi

  if [ -n "$COL1_ID" ] && [ -n "$BOARD_ID" ]; then
    ./fz column delete "$BOARD_ID" "$COL1_ID" --yes 2>/dev/null || true
  fi

  if [ -n "$COL2_ID" ] && [ -n "$BOARD_ID" ]; then
    ./fz column delete "$BOARD_ID" "$COL2_ID" --yes 2>/dev/null || true
  fi

  if [ -n "$BOARD_ID" ]; then
    ./fz board delete "$BOARD_ID" --yes 2>/dev/null || true
  fi

  echo "Cleanup complete."
}
trap cleanup EXIT

# ══════════════════════════════════════════════
# Phase 1: Auth & Identity
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 1: Auth & Identity ==="
run_test "auth status" "./fz auth status" "Authenticated"
run_test "auth token" "./fz auth token"
run_test "api /my/identity" "./fz api /my/identity" "accounts"

# ══════════════════════════════════════════════
# Phase 2: Users & Tags
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 2: Users & Tags ==="
run_test "user list" "./fz user list" "ID"

# Capture the first user ID.
USER_ID=$(./fz user list 2>/dev/null | awk 'NR==2 {print $1}')
if [ -n "$USER_ID" ]; then
  run_test "user view" "./fz user view $USER_ID" "$USER_ID"
fi

run_test "tag list" "./fz tag list"

# ══════════════════════════════════════════════
# Phase 3: Status & Raw API
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 3: Status & Raw API ==="
run_test "status" "./fz status" "Account:"
run_test "api boards (auto-slug)" "./fz api boards"

# ══════════════════════════════════════════════
# Phase 4: Board CRUD
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 4: Board CRUD ==="
run_test "board create" "./fz board create '$SUFFIX-board'" "created"

# Extract board ID via the API.
BOARD_ID=$(./fz api boards 2>/dev/null | jq -r ".[] | select(.name == \"$SUFFIX-board\") | .id")
if [ -z "$BOARD_ID" ]; then
  echo "  ${RED}FAIL${NC} could not extract board ID"
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
else
  echo "  Board ID: $BOARD_ID"
fi

run_test "board list" "./fz board list" "$SUFFIX-board"
run_test "board view (kanban)" "./fz board view $BOARD_ID"
run_test "board view --json" "./fz board view $BOARD_ID --json" "board"
run_test "board publish" "./fz board publish $BOARD_ID" "published"
run_test "board unpublish" "./fz board unpublish $BOARD_ID --yes" "unpublished"

# ══════════════════════════════════════════════
# Phase 5: Column CRUD
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 5: Column CRUD ==="
run_test "column create 1" "./fz column create $BOARD_ID --name '${SUFFIX}-col1' --color Blue" "created"
run_test "column create 2" "./fz column create $BOARD_ID --name '${SUFFIX}-col2' --color Yellow" "created"

run_test "column list" "./fz column list $BOARD_ID" "${SUFFIX}-col1"

# Extract column IDs.
COL1_ID=$(./fz column list "$BOARD_ID" 2>/dev/null | awk -v n="${SUFFIX}-col1" '$0 ~ n {print $1}')
COL2_ID=$(./fz column list "$BOARD_ID" 2>/dev/null | awk -v n="${SUFFIX}-col2" '$0 ~ n {print $1}')
echo "  Col1 ID: $COL1_ID"
echo "  Col2 ID: $COL2_ID"

# ══════════════════════════════════════════════
# Phase 6: Card CRUD
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 6: Card CRUD ==="
run_test "card create 1" "./fz card create -b $BOARD_ID -t '${SUFFIX}-card1' -B 'First test card'" "created"
run_test "card create 2" "./fz card create -b $BOARD_ID -t '${SUFFIX}-card2'" "created"

run_test "card list" "./fz card list -b $BOARD_ID" "${SUFFIX}-card1"

# Extract card numbers.
CARD1_NUM=$(./fz card list -b "$BOARD_ID" 2>/dev/null | awk -v n="${SUFFIX}-card1" '$0 ~ n {print $1}')
CARD2_NUM=$(./fz card list -b "$BOARD_ID" 2>/dev/null | awk -v n="${SUFFIX}-card2" '$0 ~ n {print $1}')
echo "  Card1 #: $CARD1_NUM"
echo "  Card2 #: $CARD2_NUM"

run_test "card view" "./fz card view $CARD1_NUM" "${SUFFIX}-card1"
run_test "card edit" "./fz card edit $CARD1_NUM -t '${SUFFIX}-card1-edited'" "updated"

# ══════════════════════════════════════════════
# Phase 7: Card Lifecycle
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 7: Card Lifecycle ==="
run_test "card triage" "./fz card triage $CARD1_NUM -c $COL1_ID" "triaged"
run_test "card untriage" "./fz card untriage $CARD1_NUM" "triage"
run_test "card close" "./fz card close $CARD1_NUM" "closed"
run_test "card reopen" "./fz card reopen $CARD1_NUM" "reopened"
run_test "card postpone" "./fz card postpone $CARD1_NUM" "Not Now"
run_test "card gild" "./fz card gild $CARD1_NUM" "goldness"
run_test "card ungild" "./fz card ungild $CARD1_NUM" "goldness removed"
run_test "card pin" "./fz card pin $CARD1_NUM" "pin"
run_test "card unpin" "./fz card unpin $CARD1_NUM" "pin removed"
run_test "card watch" "./fz card watch $CARD1_NUM" "watch"
run_test "card unwatch" "./fz card unwatch $CARD1_NUM" "watch removed"

# ══════════════════════════════════════════════
# Phase 8: Card Tagging
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 8: Card Tagging ==="
run_test "card tag" "./fz card tag $CARD1_NUM --tag '${SUFFIX}-tag'" "toggled"
run_test "card tag (untoggle)" "./fz card tag $CARD1_NUM --tag '${SUFFIX}-tag'" "toggled"

# ══════════════════════════════════════════════
# Phase 9: Steps
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 9: Steps ==="
run_test "step create" "./fz step create $CARD1_NUM --content '${SUFFIX}-step'" "Step added"

# Extract step ID via the API.
STEP_ID=$(./fz api "cards/$CARD1_NUM" 2>/dev/null | jq -r '.steps[-1].id // empty')
if [ -n "$STEP_ID" ]; then
  echo "  Step ID: $STEP_ID"
  run_test "step view" "./fz step view $CARD1_NUM $STEP_ID" "${SUFFIX}-step"
  run_test "step edit" "./fz step edit $CARD1_NUM $STEP_ID --completed" "updated"
  run_test "step delete" "./fz step delete $CARD1_NUM $STEP_ID --yes" "deleted"
fi

# ══════════════════════════════════════════════
# Phase 10: Comments
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 10: Comments ==="
run_test "comment create" "./fz comment create $CARD1_NUM --body '${SUFFIX}-comment'" "Comment added"

run_test "comment list" "./fz comment list $CARD1_NUM" "${SUFFIX}-comment"

# Extract the last non-System comment ID.
COMMENT_ID=$(./fz comment list "$CARD1_NUM" 2>/dev/null | awk -v pat="${SUFFIX}-comment" '$0 ~ pat {print $1}' | tail -1)
if [ -n "$COMMENT_ID" ]; then
  echo "  Comment ID: $COMMENT_ID"
  run_test "comment view" "./fz comment view $CARD1_NUM $COMMENT_ID" "${SUFFIX}-comment"
  run_test "comment edit" "./fz comment edit $CARD1_NUM $COMMENT_ID --body '${SUFFIX}-comment-edited'" "updated"
fi

# ══════════════════════════════════════════════
# Phase 11: Reactions
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 11: Reactions ==="

# Card reaction.
run_test "card reaction create" "./fz card reaction create $CARD1_NUM --body '👍'" "Reaction added"

REACTION_ID=$(./fz card reaction list "$CARD1_NUM" 2>/dev/null | awk 'NR==2 {print $1}')
if [ -n "$REACTION_ID" ]; then
  echo "  Card Reaction ID: $REACTION_ID"
  run_test "card reaction list" "./fz card reaction list $CARD1_NUM" "$REACTION_ID"
  run_test "card reaction delete" "./fz card reaction delete $CARD1_NUM $REACTION_ID --yes" "Reaction removed"
fi

# Comment reaction.
if [ -n "$COMMENT_ID" ]; then
  run_test "comment reaction create" "./fz comment reaction create $CARD1_NUM $COMMENT_ID --body '🎉'" "Reaction added"

  COMMENT_REACTION_ID=$(./fz comment reaction list "$CARD1_NUM" "$COMMENT_ID" 2>/dev/null | awk 'NR==2 {print $1}')
  if [ -n "$COMMENT_REACTION_ID" ]; then
    echo "  Comment Reaction ID: $COMMENT_REACTION_ID"
    run_test "comment reaction list" "./fz comment reaction list $CARD1_NUM $COMMENT_ID" "$COMMENT_REACTION_ID"
    run_test "comment reaction delete" "./fz comment reaction delete $CARD1_NUM $COMMENT_ID $COMMENT_REACTION_ID --yes" "removed"
  fi
fi

# ══════════════════════════════════════════════
# Phase 12: Webhooks
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 12: Webhooks ==="
run_test "webhook create" "./fz webhook create $BOARD_ID --name '${SUFFIX}-hook' --url 'https://example.com/hook'" "created"

run_test "webhook list" "./fz webhook list $BOARD_ID" "${SUFFIX}-hook"

WEBHOOK_ID=$(./fz webhook list "$BOARD_ID" 2>/dev/null | awk -v n="${SUFFIX}-hook" '$0 ~ n {print $1}')
if [ -n "$WEBHOOK_ID" ]; then
  echo "  Webhook ID: $WEBHOOK_ID"
  run_test "webhook view" "./fz webhook view $BOARD_ID $WEBHOOK_ID" "${SUFFIX}-hook"
  run_test "webhook edit" "./fz webhook edit $BOARD_ID $WEBHOOK_ID --name '${SUFFIX}-hook-edited'" "updated"
fi

# ══════════════════════════════════════════════
# Phase 13: Notifications
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 13: Notifications ==="
run_test "notification list" "./fz notification list"
run_test "notification read-all" "./fz notification read-all --yes" "read"

# ══════════════════════════════════════════════
# Phase 14: Card Assignment
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 14: Card Assignment ==="
if [ -n "$USER_ID" ] && [ -n "$CARD1_NUM" ]; then
  run_test "card assign" "./fz card assign $CARD1_NUM --assignee $USER_ID" "toggled"
  run_test "card unassign" "./fz card assign $CARD1_NUM --assignee $USER_ID" "toggled"
fi

# ══════════════════════════════════════════════
# Phase 15: Explicit Cleanup
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 15: Explicit Cleanup ==="

if [ -n "$WEBHOOK_ID" ] && [ -n "$BOARD_ID" ]; then
  run_test "webhook delete" "./fz webhook delete $BOARD_ID $WEBHOOK_ID --yes" "deleted"
  WEBHOOK_ID=""
fi

if [ -n "$COMMENT_ID" ] && [ -n "$CARD1_NUM" ]; then
  run_test "comment delete" "./fz comment delete $CARD1_NUM $COMMENT_ID --yes" "deleted"
  COMMENT_ID=""
fi

if [ -n "$CARD1_NUM" ]; then
  run_test "card delete 1" "./fz card delete $CARD1_NUM --yes" "deleted"
  CARD1_NUM=""
fi

if [ -n "$CARD2_NUM" ]; then
  run_test "card delete 2" "./fz card delete $CARD2_NUM --yes" "deleted"
  CARD2_NUM=""
fi

if [ -n "$COL1_ID" ] && [ -n "$BOARD_ID" ]; then
  run_test "column delete 1" "./fz column delete $BOARD_ID $COL1_ID --yes" "deleted"
  COL1_ID=""
fi

if [ -n "$COL2_ID" ] && [ -n "$BOARD_ID" ]; then
  run_test "column delete 2" "./fz column delete $BOARD_ID $COL2_ID --yes" "deleted"
  COL2_ID=""
fi

if [ -n "$BOARD_ID" ]; then
  run_test "board delete" "./fz board delete $BOARD_ID --yes" "deleted"
  BOARD_ID=""
fi

# ══════════════════════════════════════════════
# Phase 16: Verify Clean State
# ══════════════════════════════════════════════
echo ""
echo "=== Phase 16: Verify Clean State ==="
BOARD_LIST=$(./fz board list 2>&1)
if echo "$BOARD_LIST" | grep -qF "$SUFFIX-board"; then
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ${RED}FAIL${NC} board still present after delete\n"
else
  PASS=$((PASS + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ${GREEN}PASS${NC} board confirmed deleted\n"
fi

# ── Summary ──
echo ""
echo "════════════════════════════════════════════"
printf "=== Results: ${GREEN}%d passed${NC}, " "$PASS"
if [ "$FAIL" -gt 0 ]; then
  printf "${RED}%d failed${NC}, " "$FAIL"
else
  printf "%d failed, " "$FAIL"
fi
echo "$TOTAL total ==="
echo "════════════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
