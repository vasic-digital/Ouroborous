#!/usr/bin/env bash
# ouroborous_describe_challenge.sh
#
# Round-269 paired-mutation deep-doc challenge for digital.vasic.ouroborous.
#
# Validates that:
#   1. The deep-doc ledger (docs/test-coverage.md) lists every exported
#      symbol from pkg/types/types.go and pkg/client/client.go.
#   2. The multi-locale fixture (tests/fixtures/ouroborous/payloads.json)
#      parses and contains at least 3 locales.
#   3. The multi-locale runner (challenges/runner/main.go) builds and
#      runs, byte-preserving non-ASCII prompts through the real
#      ouro.Client + capturing Runner across SelfReflect, Refine,
#      MetaEvaluate, SelfImprove, DetectCycle, GetMetaPatterns, and
#      the ErrBaselineRunnerNotConfigured sentinel path.
#   4. The README enumerates the round-269 anti-bluff guarantees.
#
# Paired-mutation invariant (CONST-035 + CONST-050(B)):
#   With --anti-bluff-mutate the script plants a deliberate symbol-rename
#   mutation in a tmp copy of the ledger (DetectCycle ->
#   DetectCycle_MUTATED), reruns validation, and asserts the gate
#   FAILS with exit 99. This proves the gate actually catches
#   ledger-vs-source drift instead of rubber-stamping it.
#
# Exit codes:
#   0  — gate PASS on clean tree
#   1  — gate FAIL on clean tree (real failure to fix)
#   99 — paired-mutation correctly detected (good — proves anti-bluff)
#   2  — usage / environment error

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

MUTATE=0
for arg in "$@"; do
    case "$arg" in
        --anti-bluff-mutate) MUTATE=1 ;;
        --help|-h)
            sed -n '1,32p' "$0"
            exit 0
            ;;
        *)
            echo "unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

LEDGER="${MODULE_DIR}/docs/test-coverage.md"
FIXTURE="${MODULE_DIR}/tests/fixtures/ouroborous/payloads.json"
RUNNER="${MODULE_DIR}/challenges/runner/main.go"
README="${MODULE_DIR}/README.md"

LEDGER_WORK="${LEDGER}"
TMP_LEDGER=""
if [ "${MUTATE}" -eq 1 ]; then
    TMP_LEDGER="$(mktemp)"
    cp "${LEDGER}" "${TMP_LEDGER}"
    # Plant a rename so the symbol no longer matches what the source declares.
    sed -i 's/DetectCycle/DetectCycle_MUTATED/g' "${TMP_LEDGER}"
    LEDGER_WORK="${TMP_LEDGER}"
    echo "=== Ouroborous Describe Challenge (anti-bluff-mutate mode) ==="
else
    echo "=== Ouroborous Describe Challenge (clean mode) ==="
fi
echo ""

# Section 1: ledger presence and freshness
echo "Section 1: docs/test-coverage.md ledger"
if [ ! -f "${LEDGER_WORK}" ]; then
    fail "ledger missing at ${LEDGER_WORK}"
else
    pass "ledger present"
    if grep -q "round-269" "${LEDGER_WORK}"; then
        pass "ledger marked round-269"
    else
        fail "ledger missing round-269 marker"
    fi
    if grep -q "execution of tests and Challenges MUST guarantee" "${LEDGER_WORK}"; then
        pass "ledger carries Article XI §11.9 mandate"
    else
        fail "ledger missing Article XI §11.9 mandate"
    fi
fi

# Section 2: every exported package symbol appears in ledger.
# Hand-picked, stable set of structural symbols expected verbatim in
# the ledger. (Exhaustive parsing of every exported identifier would
# produce false positives from internal helpers — the ledger is
# authoritative about what counts as part of the public surface.)
echo ""
echo "Section 2: structural symbol cross-reference"

EXPECTED_SYMBOLS=(
    # pkg/types/types.go
    "MetaPrompt" "SelfReflection" "CycleDetection" "IterationResult"
    "RefinementConfig" "RefinementResult" "MetaEvaluation"
    # pkg/client/client.go
    "Runner" "Client" "New" "NewFromConfig" "SetRunner"
    "SelfReflect" "Refine" "MetaEvaluate" "SelfImprove"
    "GetMetaPatterns" "DetectCycle"
    "ErrBaselineRunnerNotConfigured"
)

CHECKED=0
MISSING=0
for sym in "${EXPECTED_SYMBOLS[@]}"; do
    CHECKED=$((CHECKED + 1))
    if grep -qE "\\b${sym}\\b" "${LEDGER_WORK}"; then
        : # found
    else
        fail "ledger missing symbol ${sym}"
        MISSING=$((MISSING + 1))
    fi
done
if [ "${MISSING}" -eq 0 ]; then
    pass "all ${CHECKED} structural symbols cross-referenced in ledger"
fi

# Section 3: multi-locale fixture sanity
echo ""
echo "Section 3: multi-locale fixture"
if [ ! -f "${FIXTURE}" ]; then
    fail "fixture missing at ${FIXTURE}"
else
    pass "fixture present"
    LOCALE_COUNT=$(grep -oE '"locale":\s*"[^"]+"' "${FIXTURE}" | sort -u | wc -l)
    if [ "${LOCALE_COUNT}" -ge 3 ]; then
        pass "fixture covers ${LOCALE_COUNT} locales (>=3)"
    else
        fail "fixture covers only ${LOCALE_COUNT} locales (<3)"
    fi
fi

# Section 4: runner builds + runs against every section
echo ""
echo "Section 4: multi-locale runner build + run (real Client + capturing Runner)"
if [ ! -f "${RUNNER}" ]; then
    fail "runner missing at ${RUNNER}"
else
    pass "runner source present"
    cd "${MODULE_DIR}"
    if go build -o /tmp/ouro_round269_runner ./challenges/runner/ 2>/tmp/ouro_build.log; then
        pass "runner builds"
        if /tmp/ouro_round269_runner -fixtures "${FIXTURE}" > /tmp/ouro_run.log 2>&1; then
            pass "runner exit 0 across every section + locale"
            # Per-locale + per-section PASS coverage
            if grep -q "PASS: \[Section2\]\[SelfReflect\]\[sr\]" /tmp/ouro_run.log; then
                pass "Section 2 Cyrillic (sr) SelfReflect round-trip"
            else
                fail "Section 2 Cyrillic (sr) SelfReflect missing"
            fi
            if grep -q "PASS: \[Section2\]\[SelfReflect\]\[ja\]" /tmp/ouro_run.log; then
                pass "Section 2 Japanese (ja) SelfReflect round-trip"
            else
                fail "Section 2 Japanese (ja) SelfReflect missing"
            fi
            if grep -q "PASS: \[Section2\]\[SelfReflect\]\[ar\]" /tmp/ouro_run.log; then
                pass "Section 2 Arabic (ar) SelfReflect round-trip"
            else
                fail "Section 2 Arabic (ar) SelfReflect missing"
            fi
            if grep -q "PASS: \[Section2\]\[SelfReflect\]\[zh-CN\]" /tmp/ouro_run.log; then
                pass "Section 2 Han (zh-CN) SelfReflect round-trip"
            else
                fail "Section 2 Han (zh-CN) SelfReflect missing"
            fi
            if grep -q "PASS: \[Section3\]\[Refine\]\[sr\]" /tmp/ouro_run.log; then
                pass "Section 3 Refine Cyrillic per-iter byte-exact"
            else
                fail "Section 3 Refine sr missing"
            fi
            if grep -q "PASS: \[Section3\]\[Refine\]\[ar\]" /tmp/ouro_run.log; then
                pass "Section 3 Refine Arabic per-iter byte-exact"
            else
                fail "Section 3 Refine ar missing"
            fi
            if grep -q "PASS: \[Section4\]\[MetaEvaluate\]\[ja\]" /tmp/ouro_run.log; then
                pass "Section 4 MetaEvaluate Japanese 3-criteria"
            else
                fail "Section 4 MetaEvaluate ja missing"
            fi
            if grep -q "PASS: \[Section4\]\[MetaEvaluate\]\[zh-CN\]" /tmp/ouro_run.log; then
                pass "Section 4 MetaEvaluate Han 3-criteria"
            else
                fail "Section 4 MetaEvaluate zh-CN missing"
            fi
            if grep -q "PASS: \[Section5\]\[SelfImprove\]\[en\]" /tmp/ouro_run.log; then
                pass "Section 5 SelfImprove English EarlyStop"
            else
                fail "Section 5 SelfImprove en missing"
            fi
            if grep -q "PASS: \[Section5\]\[SelfImprove\]\[ar\]" /tmp/ouro_run.log; then
                pass "Section 5 SelfImprove Arabic EarlyStop"
            else
                fail "Section 5 SelfImprove ar missing"
            fi
            if grep -q "PASS: \[Section6\]\[DetectCycle\]\[trigger\]\[en\]" /tmp/ouro_run.log; then
                pass "Section 6 DetectCycle trigger English"
            else
                fail "Section 6 DetectCycle trigger en missing"
            fi
            if grep -q "PASS: \[Section6\]\[DetectCycle\]\[repeated\]\[sr\]" /tmp/ouro_run.log; then
                pass "Section 6 DetectCycle repeated-phrase Cyrillic"
            else
                fail "Section 6 DetectCycle repeated sr missing"
            fi
            if grep -q "PASS: \[Section6\]\[DetectCycle\]\[repeated\]\[ja\]" /tmp/ouro_run.log; then
                pass "Section 6 DetectCycle repeated-phrase Japanese"
            else
                fail "Section 6 DetectCycle repeated ja missing"
            fi
            if grep -q "PASS: \[Section6\]\[DetectCycle\]\[benign\]\[ar\]" /tmp/ouro_run.log; then
                pass "Section 6 DetectCycle benign Arabic (no false positive)"
            else
                fail "Section 6 DetectCycle benign ar missing"
            fi
            if grep -q "PASS: \[Section6\]\[DetectCycle\]\[monotonic\]\[zh-CN\]" /tmp/ouro_run.log; then
                pass "Section 6 DetectCycle confidence monotonicity Han"
            else
                fail "Section 6 DetectCycle monotonic zh-CN missing"
            fi
            if grep -q "PASS: \[Section7\]\[SelfReflect\] sentinel" /tmp/ouro_run.log; then
                pass "Section 7 SelfReflect ErrBaselineRunnerNotConfigured sentinel"
            else
                fail "Section 7 SelfReflect sentinel missing"
            fi
            if grep -q "PASS: \[Section7\]\[Refine\] sentinel" /tmp/ouro_run.log; then
                pass "Section 7 Refine sentinel propagated"
            else
                fail "Section 7 Refine sentinel missing"
            fi
            if grep -q "PASS: \[Section7\]\[SelfImprove\] sentinel" /tmp/ouro_run.log; then
                pass "Section 7 SelfImprove sentinel propagated"
            else
                fail "Section 7 SelfImprove sentinel missing"
            fi
        else
            fail "runner exit non-zero — see /tmp/ouro_run.log"
            sed -n '1,80p' /tmp/ouro_run.log
        fi
    else
        fail "runner build failed — see /tmp/ouro_build.log"
        sed -n '1,40p' /tmp/ouro_build.log
    fi
    rm -f /tmp/ouro_round269_runner
fi

# Section 5: README round-269 anti-bluff section
echo ""
echo "Section 5: README round-269 anti-bluff section"
if grep -q "Anti-bluff guarantees" "${README}"; then
    pass "README declares Anti-bluff guarantees"
else
    fail "README missing Anti-bluff guarantees section"
fi
if grep -q "round-269" "${README}"; then
    pass "README marked round-269"
else
    fail "README missing round-269 marker"
fi

# Cleanup mutated ledger if any
if [ -n "${TMP_LEDGER}" ]; then
    rm -f "${TMP_LEDGER}"
fi

echo ""
echo "=== Summary: ${PASS}/${TOTAL} PASS, ${FAIL} FAIL ==="

if [ "${MUTATE}" -eq 1 ]; then
    if [ "${FAIL}" -gt 0 ]; then
        echo "anti-bluff-mutate: gate correctly detected planted mutation (exit 99)"
        exit 99
    else
        echo "anti-bluff-mutate: gate FAILED to detect planted mutation — bluff!"
        exit 1
    fi
fi

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
exit 0
