.PHONY: build run test test-race test-cover test-cover-html smoke clean new-game guess

## Build: compile the server binary to bin/wordgame
build:
	go build -o bin/wordgame ./cmd/wordgame/

## Run: start the server on localhost:1337
run:
	go run ./cmd/wordgame/

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

## New-game: hit POST /new and pretty-print the result (requires jq)
new-game:
	@RESP=$$(curl -s -w '\n%{http_code}' -X POST http://localhost:1337/new 2>&1) || { \
		echo "ERROR: Server not running. Start it with: make run"; \
		exit 1; \
	}; \
	HTTP_CODE=$$(echo "$$RESP" | tail -1); \
	if [ "$$HTTP_CODE" != "200" ]; then \
		echo "ERROR: Server returned HTTP $$HTTP_CODE. Is the server running?"; \
		exit 1; \
	fi; \
	echo "$$RESP" | sed '$$d' | jq .

## Guess: hit POST /guess with ID and GUESS vars (requires jq)
guess:
	@if [ -z "$(ID)" ]; then \
		echo "ERROR: Missing ID. Usage: make guess ID=<uuid> GUESS=<letter>"; \
		exit 1; \
	fi; \
	if [ -z "$(GUESS)" ]; then \
		echo "ERROR: Missing GUESS. Usage: make guess ID=<uuid> GUESS=<letter>"; \
		exit 1; \
	fi; \
	RESP=$$(curl -s -w '\n%{http_code}' -X POST http://localhost:1337/guess \
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

## Clean: remove build artifacts and coverage files
clean:
	rm -rf bin/ coverage.out
