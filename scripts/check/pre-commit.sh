#!/bin/sh
set -eu

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

go_cmd="${GO:-go}"

run_step() {
	name="$1"
	shift
	printf "\n==> %s\n" "$name" >&2
	"$@"
}

check_go_format() {
	files="$(git ls-files '*.go')"
	if [ -z "$files" ]; then
		return 0
	fi
	unformatted="$(gofmt -l $files)"
	if [ -n "$unformatted" ]; then
		printf "%s\n" "gofmt required:" >&2
		printf "%s\n" "$unformatted" >&2
		return 1
	fi
}

run_step "go mod tidy" "$go_cmd" mod tidy
run_step "gofmt check" check_go_format
run_step "go test" "$go_cmd" test ./...
run_step "diff whitespace check" git diff --check

printf "\npre-commit checks passed\n" >&2
