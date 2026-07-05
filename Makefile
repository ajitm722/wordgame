.PHONY: build run dev deps test test-race test-cover test-cover-html smoke fmt vet lint check clean new-game guess generate generate-js generate-go npm-install demo demo-frontend demo-frontend-record

## Build: compile the server binary to bin/wordgame (requires make generate-js first)
build:
	go build -o bin/wordgame ./cmd/wordgame/

## Deps: install tool dependencies
deps: npm-install
	@if ! command -v go-bindata >/dev/null 2>&1; then \
		echo "Installing go-bindata..."; \
		go install github.com/kevinburke/go-bindata/v4/go-bindata@latest; \
	fi

## Run: start the server on localhost:1337 (requires make generate first)
run:
	go run ./cmd/wordgame/

## Dev: one-command hot-reload — build JS, generate bindata -debug, webpack watch, go run
dev:
	@if ! command -v go-bindata >/dev/null 2>&1; then \
		echo "Installing go-bindata..."; \
		go install github.com/kevinburke/go-bindata/v4/go-bindata@latest; \
	fi
	@( \
		trap 'kill 0' SIGINT EXIT; \
		echo "Building frontend JS bundle..."; \
		npm run build && \
		echo "Generating go-bindata (debug mode)..."; \
		go-bindata \
			-debug \
			-pkg=bindata \
			-o=internal/bindata/generated.go \
			assets/... && \
		echo "Starting webpack watch (auto-rebuilds on save)..."; \
		npx webpack --mode development --watch & \
		echo "Waiting for initial webpack build..."; \
		sleep 3 && \
		echo "Starting Go server on http://localhost:1337"; \
		go run ./cmd/wordgame/; \
	)

## Test: run all tests with verbose output
test:
	go test ./... -v

## Test (race): run all tests with race detector
test-race:
	go test -race ./...

## Coverage: run tests and print per-package coverage percentages
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

## Coverage (HTML): open coverage report in browser
test-cover-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

## Smoke: run end-to-end HTTP smoke tests against a real server
smoke:
	go test -v -run '^TestSmoke' ./cmd/wordgame/

## Fmt: format all Go code
fmt:
	go fmt ./...

## Vet: run Go vet (reports suspicious constructs)
vet:
	go vet ./...

## Lint: run golangci-lint (must be installed separately: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

## Check: run all quality checks in sequence (fmt, vet, lint, test)
check: fmt vet lint test

## New-game: hit POST /api/v1/new and pretty-print the result (requires jq)
new-game:
	@RESP=$$(curl -s -w '\n%{http_code}' -X POST http://localhost:1337/api/v1/new 2>&1) || { \
		echo "ERROR: Server not running. Start it with: make run"; \
		exit 1; \
	}; \
	HTTP_CODE=$$(echo "$$RESP" | tail -1); \
	if [ "$$HTTP_CODE" != "200" ]; then \
		echo "ERROR: Server returned HTTP $$HTTP_CODE. Is the server running?"; \
		exit 1; \
	fi; \
	echo "$$RESP" | sed '$$d' | jq .

## Guess: hit POST /api/v1/guess with ID and GUESS vars (requires jq)
guess:
	@if [ -z "$(ID)" ]; then \
		echo "ERROR: Missing ID. Usage: make guess ID=<uuid> GUESS=<letter>"; \
		exit 1; \
	fi; \
	if [ -z "$(GUESS)" ]; then \
		echo "ERROR: Missing GUESS. Usage: make guess ID=<uuid> GUESS=<letter>"; \
		exit 1; \
	fi; \
	RESP=$$(curl -s -w '\n%{http_code}' -X POST http://localhost:1337/api/v1/guess \
		-H "Content-Type: application/json" \
		-d '{"id":"$(ID)","guess":"$(GUESS)"}' 2>&1) || { \
		echo "ERROR: Server not running. Start it with: make run"; \
		exit 1; \
	}; \
	HTTP_CODE=$$(echo "$$RESP" | tail -1); \
	if [ "$$HTTP_CODE" != "200" ]; then \
		echo "ERROR: Server returned HTTP $$HTTP_CODE."; \
		echo "$$RESP" | sed '$$d' | jq . 2>/dev/null || true; \
		exit 1; \
	fi; \
	echo "$$RESP" | sed '$$d' | jq .

## Generate: build JS bundle, then run go-bindata + Go compile (single make target)
generate: generate-js generate-go

## Generate JS: run webpack to produce bundle.[hash].js + index.html
generate-js:
	npm run build

## Generate Go: run go-bindata to embed assets, then compile Go binary
generate-go:
	@if ! command -v go-bindata >/dev/null 2>&1; then \
		echo "Installing go-bindata..."; \
		go install github.com/kevinburke/go-bindata/v4/go-bindata@latest; \
	fi
	go-bindata \
		-pkg=bindata \
		-o=internal/bindata/generated.go \
		assets/...
	go build -o bin/wordgame ./cmd/wordgame/

## Install frontend npm dependencies
npm-install:
	npm install

## Clean: remove build artifacts and coverage files
clean:
	rm -rf bin/ coverage.out assets/bundle.* assets/index.html node_modules/ internal/bindata/generated.go

## Demo: record the terminal/CLI demo GIF (VHS tape — backend perspective)
demo: generate
	@if ! command -v vhs >/dev/null 2>&1; then \
		echo "Installing VHS..."; \
		go install github.com/charmbracelet/vhs@latest; \
	fi
	vhs demo.tape

## Demo-frontend: record the browser demo GIF (Playwright + ffmpeg)
## One-shot. Builds the JS bundle, runs the server, drives the browser through
## a full game lifecycle (win + loss), then converts the webm to a GIF.
demo-frontend: generate
	@if ! command -v ffmpeg >/dev/null 2>&1; then \
		echo "Error: ffmpeg not found. Install with 'sudo apt install ffmpeg' or similar."; \
		exit 1; \
	fi
	@set -e; \
	echo "Starting server on :1337..."; \
	./bin/wordgame & \
	SERVER_PID=$$!; \
	trap 'kill $$SERVER_PID 2>/dev/null || true' EXIT; \
	sleep 2; \
	echo "Running Playwright demo script..."; \
	node scripts/demo-frontend.js; \
	echo "Converting webm to gif..."; \
	bash scripts/demo-frontend.sh; \
	kill $$SERVER_PID 2>/dev/null || true; \
	echo "✓ Demo-frontend complete. See docs/assets/demo-frontend.gif"

## Demo-frontend-record: same as demo-frontend, but reuses an already-running
## server on :1337 (caller is responsible for starting/stopping it).
## Useful in CI or when iterating on the script itself.
demo-frontend-record:
	@if ! command -v ffmpeg >/dev/null 2>&1; then \
		echo "Error: ffmpeg not found. Install with 'sudo apt install ffmpeg' or similar."; \
		exit 1; \
	fi
	@set -e; \
	echo "Recording against $${DEMO_URL:-http://localhost:1337}..."; \
	node scripts/demo-frontend.js; \
	bash scripts/demo-frontend.sh; \
	echo "✓ Done. See docs/assets/demo-frontend.gif"
