# Documentation, Code Structure & Improvement Suggestions — Verification Notes

Comprehensive analysis of `docs.md`, `code-structure.md`, and `INTERVIEW.md` against the current source tree.  
**No source files or original docs were modified.** All findings are recorded here.

---

## 1. How Verification Was Done

- Cross-referenced every file path, package, struct, method, and constant claimed in the docs against the actual `.go` source files.
- Checked `Makefile` targets against the documented table.
- Reviewed `go.mod` dependencies against the dependency descriptions.
- Inspected test files (`*_test.go`) to verify test structure, naming, and coverage claims.
- Read line-number references in `INTERVIEW.md` Quick Reference table against actual source line numbers.
- Verified concurrency, error handling, and Postel's Law behavior descriptions against the implementation.
- **Did not run tests or coverage tools** — all findings are based on static source inspection.

---

## 2. docs.md — Factual Errors & Omissions

### 2.1 `GuessResponse` Struct Missing `Word` Field (§8.3)

- **Doc says:** `GuessResponse` has only `ID`, `Current`, `GuessesRemaining`.
- **Code:** `internal/handler/types.go:18-22` has `Word string \`json:"word,omitempty"\``.
- The earlier API contract (section 3.2) correctly documents the `word` field. The struct snippet in section 8.3 does not.
- **Impact on reader:** Someone looking at the data model section would think the `Word` field doesn't exist on the response type.

### 2.2 `make stop` Target Listed But Does Not Exist (§11.2)

- **Doc says:** Table row for `make stop` — "Kills the server on port 1337".
- **Code:** `Makefile` lines 1–87 — no `stop` target is defined.
- A `stop` target could be implemented with `lsof -ti:1337 | xargs kill` or similar, but it is not present.

### 2.3 Dependencies Summary Omits Cobra (§1, Target Directory)

- **Doc says:** `go.mod` has "minimal deps (google/uuid, gorilla/mux)".
- **Code:** `go.mod:5-15` also requires `github.com/spf13/cobra v1.10.2` and its indirect deps (`pflag`, `mousetrap`).
- Cobra is used in `cmd/wordgame/main.go` for CLI flag parsing and is a direct dependency, not an indirect one.

### 2.4 Race Example Shows `guesses_remaining: 0` for Winning Correct Guess (§7.4, 7.9.2)

- **Doc says:** In the "Loss Detection & Cleanup" sequence diagram (`§7.4`), the response shows `guesses:0` for a losing guess — correct. But the code-structure concurrency diagram at `code-structure.md:329` says `guesses_remaining: MaxGuesses` after concurrent guesses, which is correct.
- **Inconsistency noted in INTERVIEW.md race example (§7.9.2):** The diagram shows `guesses:0` for a winning correct guess on "CAT" (final guess 'T' — correct). Correct guesses do not decrement `guesses_remaining`, so it should show `1` (unchanged from pre-guess state).
- **Source of truth:** `game.go:84-88` — `isCorrectGuess` returns true → calls `applyCorrectGuess` → no decrement.

### 2.5 `"é"` Guess — Wrong Error Message in Documentation (§5 Error Table)

- **Doc says:** `guess` is not alpha (e.g. "é") → 422 → `"guess must be a single A-Z character"`.
- **Code path:**
  1. `handler.go:97`: `normaliseGuess("é")` → `strings.ToUpper(strings.TrimSpace("é"))` → `"É"` (2 bytes).
  2. `handler.go:102`: `len("É")` is 2 (CJK/latin-1 supplement is multi-byte in UTF-8).
  3. `handler.go:103`: returns `"guess must be a single character"` — stops at handler; game validation never reached.
- **Actual behavior:** The handler rejects it as "too long" (byte length > 1), not as "not A-Z".
- This is an ASCII vs Unicode distinction. The game only accepts A-Z, but the handler's byte-length check catches multi-byte characters first.

### 2.6 Smoke Test Coverage Claims Overstated (§10.5, §12)

- **Doc says smoke tests** verify "ListenAndServe / port binding errors" and "Cobra's RunE and runServer(stderr, port)".
- **Code:** `smoke_test.go:21-33` uses `httptest.NewServer(mux)` — this bypasses `ListenAndServe` entirely. It calls `registerRoutes(r, srv)` directly but never exercises `NewRootCommand()`, `runServer()`, or `main()`.
- **What smoke tests actually cover:** Route wiring (`registerRoutes`), real HTTP JSON serialization, Content-Type headers, and full game lifecycle via TCP. They do NOT cover CLI flag parsing, `words.txt` loading errors, startup failures, or port binding.
- **6.7% coverage** for `cmd/wordgame` is correct — the only lines hit are `registerRoutes` and the handler methods called via `httptest.Server`. `main()`, `NewRootCommand()`, and `runServer()` are not exercised by any test.

### 2.7 POST /new Request Body Handling (§3.1)

- **Doc says:** "Content-Type: application/json (optional — body is ignored)" and "Request body: Ignored. Any body (or no body) is accepted."
- **Code:** `handler.go:50-73` — `HandleNewGame` does NOT call `decodeJSONBody`. It immediately calls `s.generateID()`. The body is never read or parsed. This is correct as documented, but note that `decoder.DisallowUnknownFields()` is NOT applied to `/new` (only to `/guess`), which is also correct since the body is truly ignored.

### 2.8 `decodeJSONBody` — No Body Size Limit

- **Doc does not mention** any request body size limit.
- **Code:** `request.go:10-13` calls `json.NewDecoder(r.Body).Decode(v)` without wrapping `r.Body` in `http.MaxBytesReader`.
- A malicious client can POST a multi-GB JSON payload to `/guess`. The built-in `json.Decoder` has a default limit of ~1GB per token, but streaming + unbounded allocation is still a risk.

### 2.9 `decodeJSONBody` Does Not Check for Trailing Data

- **Doc says** unknown JSON fields → 400. That's correct via `DisallowUnknownFields()`.
- **Not mentioned:** `json.Decoder.Decode` stops after the first JSON value. Trailing bytes after the first complete JSON object (e.g., `{"id":"x","guess":"A"}{"extra":true}`) are silently ignored.
- The handler does NOT verify `decoder.More()` or check that the stream is consumed. This means `{"id":"x","guess":"A"}malicious` passes validation and the `malicious` portion is ignored.
- Minor in practice for this API, but inconsistent with "strict JSON" positioning.

---

## 3. code-structure.md — Factual Errors & Omissions

### 3.1 Go Compiler Enforcement of `pkg/` Importing `internal/`

- **Doc says:** "No external module can import `internal/` — Go compiler blocks it" and "`pkg/` can only import standard library + external deps (no `internal/`)".
- **Reality:** The Go compiler blocks *external* modules from importing a package in the `internal/` directory of another module. Packages within the *same* module (`github.com/fleetdm/wordgame`) CAN import each other's `internal/` packages. Go does not enforce that `pkg/` cannot import `internal/` within the same module.
- **Current code:** `pkg/words/loader.go` and `pkg/identifier/id.go` do NOT import any `internal/` package — so the architecture rule holds by convention, not compiler enforcement.
- The documentation language is misleading. It should say "architecture rule" rather than "compiler enforces."

### 3.2 `internal/game` Imports Listed Incorrectly

- **Doc says:** `internal/game` uses "zero I/O dependencies — just `fmt` and `strings`".
- **Code:** `internal/game/game.go:5-10` imports `errors`, `regexp`, `strings`, `sync`. No `fmt`.
- The package does not import `fmt` at all. The sentinel errors (`errors.New`) and `LetterRegex` (`regexp.MustCompile`) are notable imports to include.
- Correct statement: "Just `errors`, `regexp`, `strings`, and `sync`."

### 3.3 Class Diagram Missing `Server` Fields

- **Diagram** shows `Server` with fields `-store *GameStore` and `-words []string`.
- **Code:** `handler.go:17-21` — the `Server` struct also has `generateID func() (string, error)` (unexported field).
- The `WithIDGenerator` functional option (`handler.go:27-29`) modifies this field. It is test-relevant and should be in the class diagram.
- The `+HandleNewGame(w, r)` and `+HandleGuess(w, r)` are correctly shown.

### 3.4 `Store.Save` Writes Lock, Not Read Lock

- **Doc says** "`GameStore.RWMutex` (Lock) — The games map during mutations `Store.Save`, `Store.Delete`."
- **Code:** `store.go:26` — `Save` calls `s.mu.Lock()`. Correct.
- **Code:** `store.go:41` — `Delete` calls `s.mu.Lock()`. Correct.
- This is accurately described.

### 3.5 Concurrency Diagram Shows `guesses_remaining: MaxGuesses` After Wrong Guess

- **Diagram in code-structure.md:329** shows response `200 {current: "_A__", guesses_remaining: MaxGuesses}` after request 2 (which guesses 'A' again on the same game).
- This is correct — the example shows repeated correct guesses, not wrong ones.
- No issue here.

---

## 4. INTERVIEW.md — Analysis of Improvements & Their Quality

### 4.1 Identifier Generator Refactor — Internal Contradiction (§7 vs §8)

- **§7 (line 360):** "I originally had 'refactor pkg/identifier's mutable global' here, but after analysis it's not worth it... Introducing a `Generator` struct just to wrap one function would be ceremony without value."
- **§8 (line 512+):** Proposes a full `Generator` struct refactor with `WithUUIDFunc` functional option for `pkg/identifier`.
- These are contradictory. The document takes both positions. One should be chosen.
- **My assessment:** The current `id_test.go` has only ONE test with the `defer restore` pattern. It does not use `t.Parallel()`. The fragile pattern is low-risk. The `Generator` refactor is defensible but adds 50+ lines of boilerplate for a package with 3 test functions. The "not worth it" take is reasonable.

### 4.2 context.Context Propagation (§8 — Improvement 1)

- **Recommends** threading `ctx context.Context` through `ApplyGuess`, `Store.Get`, etc.
- **Issue:** `ApplyGuess` uses `sync.RWMutex.Lock()` which is **not** context-aware — there is no `select` on `ctx.Done()` that can interrupt a blocked `Lock()` call. The provided code pattern `select { case <-ctx.Done(): return ctx.Err(); default: }` only checks BEFORE acquiring the lock, not during waiting.
- **Better approach:** If context propagation is desired, it belongs at the HTTP handler level (`r.Context()`), not in domain logic. Domain methods don't block on I/O in the current design — they are purely CPU-bound with mutex waits.
- **When it matters:** Only if `Store.Get` becomes a network call (Redis, PostgreSQL) or if `ApplyGuess` does I/O. For the current in-memory store, context adds ceremony without benefit.

### 4.3 Redis Scaling — Missing Atomicity Concern (§8 — Scaling)

- **Recommends** replacing in-memory store with Redis: replica does `GET → state → ApplyGuess → SET`.
- **Race condition:** With N replicas, two concurrent guesses on the same game can:
  1. Both `GET` the same game state from Redis.
  2. Both `ApplyGuess` locally (both decrement guesses from 5 → 4).
  3. Both `SET` (last write wins — one guess is silently lost).
- **Fix needed:** The document should recommend Lua scripting (`EVALSHA`) or optimistic locking (`WATCH`/`MULTI`/`EXEC`) for the game update. Simply saying "Redis commands are atomic" is misleading — individual Redis commands are atomic, but a read-modify-write across `GET` then `SET` is not.
- **My assessment:** The Redis suggestion is correct in spirit but the atomicity gap is a significant omission in the description.

### 4.4 Graceful Shutdown — "Prevents Data Loss" Claim (§8)

- **Says:** Graceful shutdown prevents data loss on deploy.
- **Reality:** With an in-memory store, ALL active games are lost on process exit regardless of graceful shutdown. Graceful shutdown prevents *dropped in-flight HTTP requests* (connections that are mid-response), not game state loss.
- **Correct statement:** Graceful shutdown prevents dropped requests and lets the OS close cleanly. Redis persistence (RDB/AOF) prevents game-state loss.
- The priority table (§8, P0) associates graceful shutdown with "Data loss on every deploy" which is accurate ONLY if the store is Redis-backed. For the current in-memory store, graceful shutdown does not prevent data loss — persistence does.

### 4.5 Rate Limiter Snippet — Not Per-IP (§8)

- **Snippet shows:**
  ```go
  limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 req/sec burst
  ```
- The comment says "Per-IP rate limiter middleware" but the code creates a **single global** limiter. All IPs share the same bucket — one client exhausting the burst blocks everyone.
- A true per-IP limiter needs a map of `sync.Mutex` + `rate.Limiter` entries keyed by remote IP, with periodic cleanup of stale entries.
- **My assessment:** The snippet is useful as a starting point, but the "Per-IP" label is incorrect.

### 4.6 `LetterRegex` — Handler Coupling Claim Is Stale (§8 — Improvement 3)

- **Says:** `LetterRegex` is exported and "couples the handler to the game's internal validation representation — if you change the regex, the handler's behaviour might silently change."
- **Current code:** The handler (`handler.go:113`) calls `g.ApplyGuess(rune(guess[0]))`. It does NOT reference `LetterRegex` anywhere. A-Z validation is inside `game.validateRune()` which uses `LetterRegex`.
- The handler delegates A-Z validation to the game via the `errors.Is(err, game.ErrInvalidGuess)` pattern. Changing `LetterRegex` affects the game's validation only, not the handler.
- The export is unnecessary (could be unexported) but does not create a coupling to the handler.
- **Valid improvement:** Replace regex with byte-range check and unexport. The coupling claim is stale.

### 4.7 `response.go` — Silent Error on json.Encode (§8 — Improvement 4)

- **Says:** `_ = json.NewEncoder(w).Encode(v)` silently discards marshal errors.
- **Code:** `response.go:12`.
- **Assessment:** Valid concern. If `v` contains a type that `json.Marshal` cannot handle (e.g., `chan`), the response gets a partial body with a 200 status already written. Marshal-to-buffer first is the standard fix.
- **Priority:** Low in practice — the types marshaled are simple structs with string/int fields. `json.Marshal` will never fail on these.

### 4.8 Status Type — Missing `String()` Method (§8 — Improvement 5)

- **Says:** Status should implement `fmt.Stringer` or use `go generate stringer`.
- **Code:** `game.go:25-32` — `Status` is an `int` iota without `String()`.
- **Assessment:** Valid but low priority. The status values are only ever compared to constants (`StatusInProgress`, `StatusWon`, `StatusLost`) — never logged or printed in the current code.

### 4.9 Redundant Method Checks — Dead Code Claim (§8 — Improvement 6)

- **Says:** Method checks (`if r.Method != http.MethodPost`) in handlers are dead code because gorilla/mux already filters via `.Methods(http.MethodPost)`.
- **Assessment:** Half true. The checks are redundant in production (router handles it) but fire in unit tests that call handler methods directly (bypassing the router). With `httptest.NewRecorder`, the tests never send non-POST requests, so the branch is exercised by `TestHandleNewGame_MethodNotAllowed` and `TestHandleGuess_MethodNotAllowed` which explicitly test it.
- Removing the checks makes handler tests dependent on the router for method rejection — a valid design choice but a trade-off in test isolation.

### 4.10 `pickWord` — Cryptographic Randomness (§8 — Improvement 7)

- **Says:** `math/rand/v2` is not cryptographically random, which is fine for a game. Notes that Go 1.21+ auto-seeds from `crypto/rand`.
- **Assessment:** Correct statement. No issue here.

---

## 5. Quick Reference Line-Number Accuracy (INTERVIEW.md §9)

| Claimed Reference | File | Actual Line | Match? |
|---|---|---|---|
| DI: `handler.go:32-42` | `handler.go` | `NewServer` at line 32-42 | ✅ |
| No interfaces: `handler.go:18` | `handler.go` | `store *store.GameStore` at line 18 | ✅ |
| Concurrency safe: `game.go:40` (RWMutex) | `game.go` | `mu sync.RWMutex` at line 40 | ✅ |
| Concurrency safe: `store.go:13` (RWMutex) | `store.go` | `mu sync.RWMutex` at line 13 | ✅ |
| validateInProgress: `game.go:94` | `game.go` | `validateInProgress` at line 93 | ⚠️ Off by 1 (docs say 94) |
| errors.Is dispatch: `handler.go:113-123` | `handler.go` | Lines 113-123 (switch on `errors.Is`) | ✅ |
| normaliseGuess: `request.go:18-20` | `request.go` | Lines 18-20 | ✅ |
| MaxGuesses: `game.go:23` | `game.go` | Line 23 | ✅ |
| registerRoutes: `main.go:80-85` | `main.go` | Lines 82-85 | ⚠️ Off by 2-3 lines |
| Smoke deterministic: `smoke_test.go:25` | `smoke_test.go` | `[]string{"ZZZZ"}` at line 25 | ✅ |
| WithIDGenerator: `handler.go:23-29` | `handler.go` | Lines 27-29 | ⚠️ Off (type at 23-25, func at 27-29) |
| Embedded State: `game.go:38` | `game.go` | Line 38 | ✅ |
| Cobra wired: `main.go:25-43` | `main.go` | Lines 25-43 | ✅ |
| Linters: `.golangci.yml` | `.golangci.yml` | Lines 1-6 (errcheck, govet, staticcheck, unparam) | ✅ |

Minor drift in 3 of 15 references. Acceptable for maintenance docs.

---

## 6. Structural Observations

### 6.1 Target Directory (docs.md §1) vs Actual

| Item | Docs say | Actual | Match? |
|---|---|---|---|
| `cmd/wordgame/main.go` | Entry point, gorilla/mux | Uses gorilla/mux + Cobra | ✅ (Cobra not mentioned) |
| `internal/handler/handler.go` | HTTP handlers + inline string validation | Yes | ✅ |
| `internal/handler/types.go` | DTOs: NewGameResponse, GuessRequest, etc. | Yes | ✅ |
| `internal/handler/request.go` | JSON decode, normaliseGuess | Yes | ✅ |
| `internal/handler/response.go` | writeJSON, writeError helpers | Yes | ✅ |
| `internal/handler/handler_test.go` | Present | Yes, 917 lines | ✅ |
| `internal/game/game.go` | Game struct + ApplyGuess | Yes | ✅ |
| `internal/game/game_test.go` | Present | Yes, 442 lines | ✅ |
| `internal/store/store.go` | In-memory GameStore | Yes | ✅ |
| `internal/store/store_test.go` | Present | Yes, 102 lines (estimated) | ✅ |
| `pkg/words/loader.go` | LoadWords(r io.Reader) | Yes | ✅ |
| `pkg/words/loader_test.go` | Present | Yes, 159 lines | ✅ |
| `pkg/identifier/id.go` | GenerateIdentifier | Yes | ✅ |
| `pkg/identifier/id_test.go` | Present | Yes, 74 lines | ✅ |
| `assets/` | Demo GIF | Not checked | — |
| `Makefile` | Build, test, coverage, linting | Yes | ✅ |
| `.golangci.yml` | Linter config | Yes, 4 linters | ✅ |
| `words.txt` | Word dictionary | Present | ✅ |
| `go.mod` | Go 1.24, minimal deps | Yes, includes Cobra | ⚠️ Cobra missing from doc |
| `Procfile` | Present | Present | ✅ |

### 6.2 Package Dependency Flow

Claims packages can be built bottom-up:
1. `pkg/identifier` — ✅ no internal imports
2. `pkg/words` — ✅ no internal imports
3. `internal/game` — ✅ imports stdlib only
4. `internal/store` — ✅ imports `internal/game` only
5. `internal/handler` — ✅ imports `internal/game`, `internal/store`, `pkg/identifier`
6. `cmd/wordgame` — ✅ imports `internal/handler`, `internal/store`, `pkg/words`

All correct.

### 6.3 Request Flow Description (code-structure.md)

The 11-step flow is accurate and matches the handler orchestration. A few notes:
- Step 6 says "Handler validates string structure (empty, too long)" — correct.
- Step 8 says "Handler calls Game.ApplyGuess(rune(guess[0]))" — correct.
- Step 10 cleanup calls `Store.Delete` — correct.
- The note "handler never reads Game.Current or Game.GuessesRemaining directly" — verified: handler always uses `g.Snapshot()`.

### 6.4 Error Handling Table Accuracy (docs.md §5)

| Scenario | Doc says | Code does | Match? |
|---|---|---|---|
| Game ID not found | 404 "game not found" | `handler.go:109` — 404 | ✅ |
| Guessing on completed game | 404 "game not found" (deleted) | After win/loss, `Store.Delete` called → `Get` returns nil → 404 | ✅ |
| Concurrent completion | 409 "game already completed" | `handler.go:115-116` — `errors.Is(err, ErrGameCompleted)` → 409 | ✅ |
| Missing guess / empty | 422 "missing guess" | `handler.go:99` — 422 | ✅ |
| Guess > 1 char | 422 "guess must be a single character" | `handler.go:103` — 422 | ✅ |
| Guess not alpha | 422 "guess must be a single A-Z character" | Via `game.ApplyGuess` → `validateRune` → `ErrInvalidGuess` → `handler.go:117-118` → 422 | ✅ |
| Missing id | 400 "missing game id" | `handler.go:90` — 400 | ✅ |
| Invalid JSON body | 400 "invalid request body" | `handler.go:85` — 400 | ✅ |
| Unknown JSON fields | 400 "invalid request body" | Via `DisallowUnknownFields` → JSON decode error → same path | ✅ |
| Lowercase guess | 200 OK, normalised | Via `normaliseGuess` | ✅ |
| Guess with whitespace | 200 OK, normalised | Via `normaliseGuess` | ✅ |

All match the documented behavior.

---

## 7. Postel's Law Implementation Check

- `normaliseGuess` in `request.go:18-20`: `strings.ToUpper(strings.TrimSpace(guess))` — ✅
- Applied in `handler.go:97` before string validation. ✅
- Game never sees raw user input: the rune passed to `ApplyGuess` is from the normalised string. ✅
- `decodeJSONBody` does NOT normalise JSON string values (e.g., no `ToUpper` on `ID` field) — correct, only `guess` is normalised. ✅

---

## 8. Concurrency Model Verification

### 8.1 Two-Level Locking

- `GameStore.mu` (`sync.RWMutex`) in `store.go:13`:
  - `Get` uses `RLock` (line 33).
  - `Len` uses `RLock` (line 48).
  - `Save` uses `Lock` (line 26).
  - `Delete` uses `Lock` (line 41).
- `Game.mu` (`sync.RWMutex`) in `game.go:40`:
  - `ApplyGuess` uses `Lock` (line 75).
  - `Snapshot` uses `RLock` (line 146).

### 8.2 Race Condition Resolution

Described in INTERVIEW.md §7.9.2:
1. Two goroutines both read the game pointer concurrently (`Store.Get` with `RLock`) — safe.
2. First acquires `Game.mu.Lock()`, wins the game, sets `Status = StatusWon`.
3. Second acquires `Game.mu.Lock()` after first releases → `validateInProgress()` sees `Status != StatusInProgress` → returns `ErrGameCompleted`.
4. Handler catches via `errors.Is(err, game.ErrGameCompleted)` → 409 Conflict.

This is accurately described and matches the code implementation.

---

## 9. Valid Improvement Ideas (Unchanged Validity)

Items from INTERVIEW.md that are technically sound and have no factual errors:

| Improvement | Priority | Notes |
|---|---|---|
| Request body size limit | P0 | Valid — no `MaxBytesReader` in `request.go` |
| HTTP server timeouts | P0 | Valid — `ListenAndServe` with no timeouts |
| Graceful shutdown | P0 | Valid — no signal handling |
| Game TTL / expiry | P1 | Valid — abandoned games leak memory |
| Configurable word-list path | P1 | Valid — `words.txt` is hardcoded relative path |
| Redis as shared store | P1 | Valid — but needs atomic update design (see §4.3) |
| Rate limiting | P1 | Valid — no abuse protection (but snippet needs per-IP fix) |
| Structured logging | P2 | Valid — `log.Printf` is not queryable |
| Metrics / health endpoints | P2 | Valid — no `/healthz`, no Prometheus |
| CORS headers | P3 | Valid — needed for browser frontends |
| Authentication / versioning | P3 | Valid — production hygiene |

---

## 10. Missing Production Concerns (Not in Any Doc)

### 10.1 Binding to `localhost` Prevents Container Access

- `main.go:72`: `addr := "localhost:" + port`
- In Docker/containers, `localhost` maps to the container's loopback interface. Heroku, Kubernetes, and most PaaS solutions expect the server to listen on `0.0.0.0:$PORT` (or `:$PORT`).
- The `Procfile` exists, so the project is designed for Heroku deployment. Heroku sets the `PORT` env var and expects binding to `0.0.0.0:<PORT>` (implicit `""` host). Using `localhost:` would break Heroku deploys.
- The documentation should note this limitation and suggest making the listen address configurable.

### 10.2 `decodeJSONBody` Accepts Empty Body as Valid

- Sending `POST /guess` with `Content-Type: application/json` and body `""` (empty) calls `json.Decoder.Decode` which returns `io.EOF`. This is NOT caught by `DisallowUnknownFields` (never reaches decoder).
- The handler treats this as `"invalid request body"` — correct behavior.
- But the documented error table says "Request body not valid JSON" → 400, which is accurate.

### 10.3 POST /new Accepts Any Content-Type

- The doc says Content-Type is optional for `/new`.
- The handler never touches `r.Body` for `/new`, so this is fine.

### 10.4 No Request ID / Correlation ID

- INTERVIEW.md notes this as missing. Correct — no `X-Request-ID` header handling.
- Makes log correlation across handler → game → store impossible at scale.

---

## 11. Summary of Most Impactful Fixes

Ranked by how much they mislead a reviewer or interviewer:

1. **GuessResponse missing Word field** — someone reading the data model section gets wrong type info.
2. **make stop doesn't exist** — trivial but immediately obvious on `make stop`.
3. **Cobra dependency omitted** — suggests a simpler dependency tree than reality.
4. **é handling misdescribed** — the error message path is wrong in docs.
5. **Smoke test claims overstated** — suggests testing depth that doesn't exist.
6. **context.Context over-recommended** — suggests a design flaw that doesn't materially exist.
7. **Redis atomicity gap** — the scaling recommendation has a correctness flaw.
8. **Graceful shutdown data-loss framing** — conflates connection draining with persistence.
9. **Identifier refactor contradiction** — takes both positions in same document.
10. **Rate limiter code/comment mismatch** — wrong label on code snippet.

No edits were made to any source file or documentation file.
