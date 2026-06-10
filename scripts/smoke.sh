#!/usr/bin/env bash
# Smoke-tests the documented CLI examples and exit codes against a fresh build.
set -euo pipefail

cd "$(dirname "$0")/.."

bin="$(mktemp -d)/lexical-cli"
go build -o "$bin" ./cmd/lexical-cli

pass=0
fail=0

check() {
  local desc="$1" want_code="$2" got_code="$3"
  if [[ "$want_code" == "$got_code" ]]; then
    echo "ok   - $desc (exit $got_code)"
    pass=$((pass + 1))
  else
    echo "FAIL - $desc (want exit $want_code, got $got_code)"
    fail=$((fail + 1))
  fi
}

run_code() { "$@" >/dev/null 2>&1; echo $?; }

# 1. file -> stdout
out="$("$bin" testdata/paragraph.json)"
[[ "$out" == "Hello world." ]] && check "file to stdout content" 0 0 || check "file to stdout content" 0 1

# 2. stdin
out="$(cat testdata/links.json | "$bin")"
[[ "$out" == *"[Lexical](https://lexical.dev)"* ]] && check "stdin pipe content" 0 0 || check "stdin pipe content" 0 1

# 3. --output writes a file
tmp="$(mktemp)"
"$bin" --output "$tmp" testdata/lists.json
grep -q "Nested one" "$tmp" && check "--output writes file" 0 0 || check "--output writes file" 0 1

# 4. formats
check "--format html" 0 "$(run_code "$bin" --format html testdata/formats.json)"
check "--format text" 0 "$(run_code "$bin" --format text testdata/formats.json)"

# 5. exit codes
check "strict on unsupported -> 1" 1 "$(run_code "$bin" --strict testdata/unsupported.json)"
check "invalid json -> 1" 1 "$(echo '{bad' | run_code "$bin")"
check "missing root -> 1" 1 "$(echo '{}' | run_code "$bin")"
check "unknown flag -> 2" 2 "$(run_code "$bin" --nope testdata/paragraph.json)"
check "bad format -> 2" 2 "$(run_code "$bin" --format yaml testdata/paragraph.json)"
check "tui non-tty -> 1" 1 "$(run_code "$bin" tui testdata/paragraph.json </dev/null)"

# 6. info commands
check "--version" 0 "$(run_code "$bin" --version)"
check "--help" 0 "$(run_code "$bin" --help)"
check "--dump-tree" 0 "$(run_code "$bin" --dump-tree testdata/formats.json)"
check "--list-unsupported" 0 "$(run_code "$bin" --list-unsupported testdata/unsupported.json)"

echo
echo "passed: $pass, failed: $fail"
[[ "$fail" -eq 0 ]]
