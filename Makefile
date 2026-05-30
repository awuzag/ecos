GO ?= go

.PHONY: help fmt-check test pre-commit verify tidy

help:
	@printf "%s\n" "ecos make targets"
	@printf "%s\n" "  make tidy       Tidy Go modules"
	@printf "%s\n" "  make fmt-check  Check Go formatting"
	@printf "%s\n" "  make test       Run Go tests"
	@printf "%s\n" "  make verify     Run all repo checks"

tidy:
	$(GO) mod tidy

fmt-check:
	@files="$$(git ls-files '*.go')"; \
	if [ -n "$$files" ]; then \
		unformatted="$$(gofmt -l $$files)"; \
		if [ -n "$$unformatted" ]; then \
			printf "%s\n" "gofmt required:"; \
			printf "%s\n" "$$unformatted"; \
			exit 1; \
		fi; \
	fi

test:
	$(GO) test ./...

pre-commit:
	scripts/check/pre-commit.sh

verify: pre-commit
