#!/usr/bin/env bash
# compare_benchmarks.sh
#
# Usage:
#   ./compare_benchmarks.sh <commit1> <commit2>
#
# This script will:
#  - require a clean git working tree
#  - remember the current HEAD/branch and restore it at exit
#  - checkout commit1, run benchmarks for package ./audio and save output to audio/bench_results/bench_<short>.txt
#  - checkout commit2, run benchmarks and save output similarly
#  - compare the two results with benchstat (if available) and save benchstat output to audio/bench_results/benchstat_<short1>_vs_<short2>.txt
#
# Notes:
#  - The script runs `go test ./audio -bench . -benchmem -run '^$' -count N`. You can tweak COUNT below.
#  - benchstat is not required but recommended for comparison. If benchstat is missing, the script will still run the benches.
#  - You must run this script from the repository root (where the .git directory is).
#
# Author: generated helper script

set -euo pipefail

BENCHTIME=20             # number of runs to average (go test -count). Increase for more stable results.
RESULT_DIR="audio/bench_results"

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <commit1> <commit2>"
    exit 2
fi

COMMIT_A="$1"
COMMIT_B="$2"

# Ensure we're inside a git repository
if [ ! -d .git ]; then
    echo "ERROR: This script must be run from the repository root (where .git is located)."
    exit 1
fi

# Require a clean working tree to avoid losing work when checking out commits
if [ -n "$(git status --porcelain)" ]; then
    echo "ERROR: Working tree is not clean. Commit or stash your changes before running this script."
    git status --porcelain
    exit 1
fi

# Remember current branch/commit so we can restore later
ORIG_BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
ORIG_COMMIT="$(git rev-parse HEAD)"

restore_original() {
    echo
    echo "Restoring original state..."
    # If we were on a named branch, check it out, else checkout the original commit
    if [ -n "$ORIG_BRANCH" ] && [ "$ORIG_BRANCH" != "HEAD" ]; then
        echo "Checking out branch $ORIG_BRANCH"
        git checkout "$ORIG_BRANCH"
    else
        echo "Checking out original commit $ORIG_COMMIT (detached HEAD)"
        git checkout "$ORIG_COMMIT"
    fi
}
trap restore_original EXIT

# Make sure bench_results dir exists
mkdir -p "$RESULT_DIR"

# Helper to run bench for one commit
run_bench_for_commit() {
    local commit="$1"

    # resolve short hash for filenames
    local short
    short="$(git rev-parse --short "$commit")" || {
        echo "ERROR: Invalid commit: $commit"
        exit 1
    }

    local out="$RESULT_DIR/bench_${short}.txt"
    echo
    echo "========================================"
    echo "Checking out $commit (short: $short)..."
    git checkout "$commit"

    echo "Running benchmarks for ./audio (commit $short) -> $out"
    echo "(This runs: go test ./audio -bench . -benchmem -run '^$' -count $COUNT )"
    # Run bench; capture both stdout and stderr
    if ! go test ./audio -bench . -benchmem -run '^$' -benchtime "$BENCHTIME" > "$out" 2>&1; then
        echo "ERROR: go test failed for commit $commit. See $out for details."
        # keep the checked-out state so the user can inspect, then exit
        exit 1
    fi

    echo "Benchmarks saved to $out"
    printf "%s\n" "$out"
}

echo "Starting benchmark comparison: $COMMIT_A vs $COMMIT_B"
echo "Original branch/commit: ${ORIG_BRANCH:-(none)} at $ORIG_COMMIT"

# Run benches for both commits
OUT_A="$(run_bench_for_commit "$COMMIT_A")"
OUT_B="$(run_bench_for_commit "$COMMIT_B")"

# Resolve short names for final filenames
SHORT_A="$(basename "$OUT_A" | sed -E 's/^bench_([a-f0-9]+)\.txt$/\1/' )"
SHORT_B="$(basename "$OUT_B" | sed -E 's/^bench_([a-f0-9]+)\.txt$/\1/' )"

# If benchstat is available, compare results
BENCHSTAT_BIN=""
if command -v benchstat >/dev/null 2>&1; then
    BENCHSTAT_BIN="benchstat"
elif command -v benchstat.exe >/dev/null 2>&1; then
    BENCHSTAT_BIN="benchstat.exe"
fi

COMPARE_OUT="$RESULT_DIR/benchstat_${SHORT_A}_vs_${SHORT_B}.txt"

if [ -n "$BENCHSTAT_BIN" ]; then
    echo
    echo "Comparing with benchstat and saving to $COMPARE_OUT"
    # benchstat expects file order (old new) typically; we treat COMMIT_A as old and COMMIT_B as new
    "$BENCHSTAT_BIN" "$OUT_A" "$OUT_B" > "$COMPARE_OUT"
    echo "benchstat output written to $COMPARE_OUT"
else
    echo
    echo "Note: benchstat not found on PATH. To compare results install benchstat (golang.org/x/perf/cmd/benchstat)."
    echo "You can compare files manually:"
    echo "  $OUT_A"
    echo "  $OUT_B"
    echo
    echo "If you install benchstat, re-run:"
    echo "  benchstat $OUT_A $OUT_B > $COMPARE_OUT"
fi

echo
echo "Done. Results are in the directory: $RESULT_DIR"
echo "Files produced:"
ls -1 "$RESULT_DIR" | sed -e 's/^/  /'

# exit normally; trap will restore original HEAD/branch
exit 0
