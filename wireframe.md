# Word Game — Frontend Wireframes & State Design

> **⚠️ Disclaimer:** This document was vibe-coded as a design artifact — it simulates the frontend experience and wireframe-first thinking for this API. The actual UI has not been implemented. Treat this as a napkin sketch that bridges the gap between "what the API returns" and "what the player sees."

---

## Why Wireframe-First?

Wireframing (usually as part of what Fleet calls "drafting") provides a clear overview of page layout, information architecture, user flow, and functionality. The wireframe-first approach extends beyond what users see on their screens — it's also excellent for drafting APIs, config settings, CLI options, and even business processes.

It's design thinking, applied to software development.

Key principles:
- We create a wireframe for every change and favor small, iterative changes to deliver value quickly.
- We can think through functionality and UX deeply before committing any code. As a result, coding decisions are clearer, and code is cleaner and easier to maintain.
- Content hierarchy, messaging, error states, interactions, URLs, API parameters, and API response data are all considered during wireframing.
- Designing from the "outside, in" lets us obsess over interaction details. An undefined "what" exposes results to chaos.
- Much like Pixar's storyboarding process, wireframing lets us inexpensively storyboard a user journey before locking in decisions that are prohibitively expensive to change post-production.

---

## Architecture: How the Frontend Talks to the Backend

```
┌──────────────────────────────────────────┐
│              PLAYER'S BROWSER             │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │         UI (React Component)        │  │
│  │  ┌──────────────────────────────┐  │  │
│  │  │      Board Display            │  │  │
│  │  │   [ _  _  P  P  _  E ]       │  │  │
│  │  │   Guesses remaining: 4       │  │  │
│  │  └──────────────────────────────┘  │  │
│  │  ┌──────────────────────────────┐  │  │
│  │  │      Letter Buttons           │  │  │
│  │  │ [A] [B] [C] [D] [E] [F] [G]  │  │  │
│  │  │ [H] [I] [J] [K] [L] [M] [N]  │  │  │
│  │  │ [O] [P] [Q] [R] [S] [T] [U]  │  │  │
│  │  │ [V] [W] [X] [Y] [Z]          │  │  │
│  │  └──────────────────────────────┘  │  │
│  │  [ New Game ]  [ Give Up ]         │  │
│  └────────────────┬───────────────────┘  │
│                   │                       │
│  ┌────────────────▼───────────────────┐  │
│  │          State (useState)           │  │
│  │  {                                  │  │
│  │    gameId: "abc123...",  ← stored   │  │
│  │    current: "_PP__E",               │  │
│  │    guesses: 4,                      │  │
│  │    status: "playing",               │  │
│  │    guessedLetters: ["P","E","Z"]    │  │
│  │  }                                  │  │
│  └────────────────┬───────────────────┘  │
│                   │                       │
│  ┌────────────────▼───────────────────┐  │
│  │        API Layer (fetch)            │  │
│  │  POST /new                          │  │
│  │  POST /guess {id, guess}           │  │
│  └────────────────┬───────────────────┘  │
└───────────────────┼──────────────────────┘
                    │  HTTP (JSON)
                    ▼
┌──────────────────────────────────────────┐
│           BACKEND (Go Server)             │
│  POST /new   → {id, current, guesses}     │
│  POST /guess → {id, current, guesses}     │
│                                          │
│  State: in-memory map[id]*Game            │
└──────────────────────────────────────────┘
```

---

## State Management: How the Game ID Flows

The frontend is the **session owner**. The UUID lives in the component, never leaked to the user. Every API call carries it under the hood.

```
┌─────────────────────────────────────────────────────────┐
│                    STATE LIFECYCLE                       │
│                                                         │
│  ╔══════════════════════════════════════════════════╗    │
│  ║  Component Mount                                  ║   │
│  ║  state = { gameId: null, status: "idle" }        ║   │
│  ╚════════════════════╤═════════════════════════════╝    │
│                       │                                   │
│              Player clicks [New Game]                     │
│                       │                                   │
│                       ▼                                   │
│  ╔══════════════════════════════════════════════════╗    │
│  ║  POST /new                                        ║   │
│  ║  ← {id:"abc123", current:"_____", guesses:6}     ║   │
│  ║                                                   ║   │
│  ║  state = {                                        ║   │
│  ║    gameId: "abc123",     ← stored, never shown    ║   │
│  ║    current: "_____",                              ║   │
│  ║    guesses: 6,                                    ║   │
│  ║    status: "playing"                              ║   │
│  ║  }                                                ║   │
│  ╚════════════════════╤═════════════════════════════╝    │
│                       │                                   │
│           Player taps letter [P]                          │
│                       │                                   │
│                       ▼                                   │
│  ╔══════════════════════════════════════════════════╗    │
│  ║  POST /guess {id:"abc123", guess:"P"}            ║   │
│  ║  ← {id:"abc123", current:"_PP__", guesses:6}    ║   │
│  ║                                                   ║   │
│  ║  state = {                                        ║   │
│  ║    gameId: "abc123",     ← same ID reused         ║   │
│  ║    current: "_PP__",     ← updated board          ║   │
│  ║    guesses: 6,                                    ║   │
│  ║    status: "playing"                              ║   │
│  ║  }                                                ║   │
│  ╚══════════════════════════════════════════════════╝    │
│                                                         │
│  ... repeat for each letter guess ...                   │
│                                                         │
│                       │                                   │
│              All letters revealed ──→ WIN                 │
│                  or guesses hit 0 ──→ LOSS               │
│                       │                                   │
│                       ▼                                   │
│  ╔══════════════════════════════════════════════════╗    │
│  ║  state = {                                        ║   │
│  ║    status: "won" | "lost"                        ║   │
│  ║  }                                                ║   │
│  ║  ← Letter buttons disabled                        ║   │
│  ║  ← [Play Again] button shown                     ║   │
│  ╚══════════════════════════════════════════════════╝    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Key detail:** The `gameId` is NEVER in the URL, NEVER visible to the player. It's just the "invisible session token" that the frontend manages entirely in memory. If the user refreshes the page, the game is lost — same as closing a physical hangman notepad. (In production, you'd persist it to `localStorage`.)

---

## Screen 1: Idle State (Before Game Starts)

```
┌─────────────────────────────────────────┐
│                                         │
│            ╔═══════════════╗            │
│            ║  WORD GAME    ║            │
│            ╚═══════════════╝            │
│                                         │
│    Guess the hidden word before         │
│    you run out of attempts.             │
│    You have 6 guesses.                  │
│                                         │
│         ┌──────────────┐               │
│         │  NEW GAME     │               │
│         └──────────────┘               │
│                                         │
│                                         │
└─────────────────────────────────────────┘

API:    none (idle)
State:  { gameId: null, status: "idle" }
```

---

## Screen 2: New Game Created — Initial Board

```
┌─────────────────────────────────────────┐
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ _ │ │ _ │ │ _ │ │ _ │ │ _ │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│                                         │
│   Guesses left:  ● ● ● ● ● ●  (6)     │
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ A │ │ B │ │ C │ │ D │ │ E │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ F │ │ G │ │ H │ │ I │ │ J │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ K │ │ L │ │ M │ │ N │ │ O │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ P │ │ Q │ │ R │ │ S │ │ T │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ U │ │ V │ │ W │ │ X │ │ Y │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│           ┌───┐                        │
│           │ Z │  [ Give Up ]           │
│           └───┘                        │
│                                         │
└─────────────────────────────────────────┘

API:    POST /new
State:  { gameId:"abc123...", current:"_____", guesses:6, status:"playing" }
```

---

## Screen 3: Correct Guess — Letter Revealed

Player taps `P`. The letter is in the word (APPLE).

```
┌─────────────────────────────────────────┐
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ _ │ │ P │ │ P │ │ _ │ │ _ │  ←!   │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│                                         │
│   Guesses left:  ● ● ● ● ● ●  (6)     │
│   "P" is correct! ✓                     │
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ A │ │ B │ │ C │ │ D │ │ E │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ F │ │ G │ │ H │ │ I │ │ J │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ K │ │ L │ │ M │ │ N │ │ O │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐      ┌───┐ ┌───┐ ┌───┐      │
│   │▓P▓│ ← dim│ Q │ │ R │ │ S │      │
│   └───┘      └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ T │ │ U │ │ V │ │ W │ │ X │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│           ┌───┐                        │
│           │ Y │ │ Z │                  │
│           └───┘ └───┘                  │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"abc123","guess":"P"}
        ← 200 {current:"_PP__", guesses_remaining:6}

State:  { gameId:"abc123", current:"_PP__", guesses:6,
          guessedLetters:["P"], status:"playing" }

Note:   'P' button grays out (or gets a ✓) — frontend tracks which
        letters were guessed locally. Duplicate taps → no API call.
```

---

## Screen 4: Wrong Guess — Penalty Applied

Player taps `Z`. The letter is NOT in the word.

```
┌─────────────────────────────────────────┐
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ _ │ │ P │ │ P │ │ _ │ │ _ │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│                                         │
│   Guesses left:  ● ● ● ● ● ○  (5) ←   │
│   "Z" is not in the word ✗              │
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ A │ │ B │ │ C │ │ D │ │ E │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ F │ │ G │ │ H │ │ I │ │ J │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ K │ │ L │ │ M │ │ N │ │ O │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│   ┌───┐      ┌───┐ ┌───┐ ┌───┐      │
│   │▓P▓│      │ Q │ │ R │ │ S │      │
│   └───┘      └───┘ └───┘ └───┘      │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ T │ │ U │ │ V │ │ W │ │ X │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│         ┌───┐ ┌───┐                    │
│         │▓Z▓│ │ Y │                    │
│         └───┘ └───┘                    │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"abc123","guess":"Z"}
        ← 200 {current:"_PP__", guesses_remaining:5}

State:  { gameId:"abc123", current:"_PP__", guesses:5,
          guessedLetters:["P","Z"], status:"playing" }

Note:   Guess counter drops. One dot goes from filled ● to empty ○.
        'Z' grays out with an ✗ indicator. The board stays unchanged.
```

---

## Screen 5: Win State

Player guessed all letters. Last guess was `E`.

```
┌─────────────────────────────────────────┐
│                                         │
│          ╔═══════════════════╗          │
│          ║   🎉 YOU WIN!    ║          │
│          ╚═══════════════════╝          │
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ A │ │ P │ │ P │ │ L │ │ E │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│            The word was: APPLE          │
│                                         │
│   Guesses left:  ● ● ● ● ○ ○  (2)     │
│                                         │
│         ┌──────────────┐               │
│         │  PLAY AGAIN   │               │
│         └──────────────┘               │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"abc123","guess":"E"}
        ← 200 {current:"APPLE", guesses_remaining:2}

State:  { gameId:"abc123", current:"APPLE", guesses:2,
          status:"won" }

Note:   All letters revealed — no underscores left in current.
        All letter buttons disabled. [Play Again] starts a new game.
        Any further guesses → 400 (game already completed).
```

---

## Screen 6: Loss State

Player exhausted all 6 guesses without finding the word (`APPLE`).

```
┌─────────────────────────────────────────┐
│                                         │
│          ╔═══════════════════╗          │
│          ║   💀 YOU LOSE    ║          │
│          ╚═══════════════════╝          │
│                                         │
│   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│   │ _ │ │ P │ │ P │ │ _ │ │ _ │      │
│   └───┘ └───┘ └───┘ └───┘ └───┘      │
│          The word was: APPLE            │
│                                         │
│   Guesses left:  ○ ○ ○ ○ ○ ○  (0)     │
│                                         │
│   Your guesses: Z, Q, X, W, V, Y       │
│                                         │
│         ┌──────────────┐               │
│         │  TRY AGAIN    │               │
│         └──────────────┘               │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"abc123","guess":"Y"}
        ← 200 {current:"_PP__", guesses_remaining:0}

State:  { gameId:"abc123", current:"_PP__", guesses:0,
          status:"lost" }

Note:   Server returns guesses_remaining:0. Frontend detects loss.
        The secret word is NOT revealed by the API — the frontend
        could show it by looking at which letters are already on the board,
        but the API spec doesn't expose the word (by design — it's hidden).
        In practice, the UI could just say "the word was: _PP__" and the
        player knows they were close.
```

---

## Screen 7: Error States

### 7a. Game Expired / Not Found (400)

```
┌─────────────────────────────────────────┐
│                                         │
│         ⚠️  Game not found             │
│                                         │
│    This game may have expired or        │
│    the server was restarted.            │
│                                         │
│         ┌──────────────┐               │
│         │  NEW GAME     │               │
│         └──────────────┘               │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"expired-id","guess":"A"}
        ← 400 {"error":"game not found"}

State:  { gameId:"expired-id", status:"error", error:"game not found" }
```

### 7b. Game Already Completed (400)

```
┌─────────────────────────────────────────┐
│                                         │
│     ⚠️  This game is already over      │
│                                         │
│         ┌──────────────┐               │
│         │  PLAY AGAIN   │               │
│         └──────────────┘               │
│                                         │
└─────────────────────────────────────────┘

API:    POST /guess {id:"abc123","guess":"X"}
        ← 400 {"error":"game already completed"}

State:  { status:"error", error:"game already completed" }
```

### 7c. Network Error (fetch failure)

```
┌─────────────────────────────────────────┐
│                                         │
│         ⚠️  Connection lost            │
│                                         │
│    Could not reach the server.          │
│    Check your connection and try again. │
│                                         │
│         ┌──────────────┐               │
│         │  RETRY        │               │
│         └──────────────┘               │
│                                         │
└─────────────────────────────────────────┘

API:    fetch() rejects (network error)
State:  { status:"error", error:"network" }
```

---

## Component Tree

```
<App>
  ├── <Header>
  │     └── "Word Game" title
  │
  ├── <StatusBar>
  │     ├── GuessesRemaining (dots: ●●●●●○)
  │     └── StatusMessage ("P is correct!" / "Z is not in the word")
  │
  ├── <Board>
  │     └── <Tile /> × wordLength
  │           └── letter or "_"
  │
  ├── <Keyboard>
  │     └── <LetterButton /> × 26
  │           ├── letter (A-Z)
  │           ├── state: "idle" | "correct" | "wrong"
  │           ├── disabled: true/false
  │           └── onClick: () => makeGuess(letter)
  │
  ├── <ActionBar>
  │     ├── [New Game]    (always visible)
  │     └── [Give Up]     (only during active game)
  │
  └── <ErrorBanner />     (conditional — game not found, network error)
```

---

## Data Flow: One Guess, End to End

```
┌──────────┐    ┌──────────────┐    ┌──────────┐    ┌──────────┐
│ Keyboard  │    │  App State   │    │ API Layer │    │ Backend  │
│ Component │    │  (useState)  │    │ (fetch)   │    │ (Go)     │
└────┬─────┘    └──────┬───────┘    └────┬─────┘    └────┬─────┘
     │                 │                 │               │
     │ onClick('P')    │                 │               │
     │────────────────→│                 │               │
     │                 │                 │               │
     │                 │ POST /guess     │               │
     │                 │ {id, guess:"P"} │               │
     │                 │────────────────→│               │
     │                 │                 │               │
     │                 │                 │  HTTP POST    │
     │                 │                 │──────────────→│
     │                 │                 │               │
     │                 │                 │   200 OK      │
     │                 │                 │  {current:    │
     │                 │                 │   "_PP__",    │
     │                 │                 │   guesses:6}  │
     │                 │                 │←──────────────│
     │                 │                 │               │
     │                 │   response      │               │
     │                 │←────────────────│               │
     │                 │                 │               │
     │                 │ setState({      │               │
     │                 │   current:      │               │
     │                 │   "_PP__",      │               │
     │                 │   guesses: 6    │               │
     │                 │ })              │               │
     │                 │                 │               │
     │  re-render      │                 │               │
     │←────────────────│                 │               │
     │  Board shows    │                 │               │
     │  "_PP__"        │                 │               │
     │  Button 'P'     │                 │               │
     │  grays out      │                 │               │
     │                 │                 │               │
```

---

## API Mapping: Wireframe State ↔ API Responses

| Wireframe Element | API Field | Source |
|-------------------|-----------|--------|
| Board tiles (`_`, `P`, `P`, `_`, `_`) | `current` string | `POST /new` and `POST /guess` responses |
| Guess dots (●●●●●○) | `guesses_remaining` int | Both responses |
| Game identifier (invisible) | `id` string | `POST /new` response |
| Letter sent to server | `guess` string | `POST /guess` request body |
| Error message | `error` string | Error responses (400) |
| "Letter is correct" indicator | Determined by frontend | Compare old `current` vs new `current` — if changed, guess was correct |
| "Letter is wrong" indicator | Determined by frontend | Compare `guesses_remaining` — if decremented, guess was wrong |

**Important:** The API never tells you "that was correct" or "that was wrong." The frontend deduces it by comparing the response to the previous state. If `current` changed → correct. If `guesses_remaining` decreased → wrong. This is a clean API separation — the backend returns data, the frontend interprets it.

---

## Responsive States (320px width)

The Fleet challenge specifically asks for responsiveness down to 320px:

```
Desktop (>768px)                  Mobile (320px)
┌──────────────────────┐          ┌──────────┐
│  BOARD: _ P P _ _    │          │ _ P P _ _│
│  Guesses: ●●●●●○     │          │ ●●●●●○   │
│                      │          │          │
│  A B C D E F G       │          │ A B C D E│
│  H I J K L M N       │          │ F G H I J│
│  O P Q R S T U       │          │ K L M N O│
│  V W X Y Z           │          │ P Q R S T│
│                      │          │ U V W X Y│
│  [New Game] [Give Up]│          │ Z        │
└──────────────────────┘          │[New Game]│
                                  └──────────┘

Changes at 320px:
  • Board tiles shrink to fit
  • Keyboard: 5 columns instead of 7
  • Guess counter: smaller dots
  • Buttons: full-width stacked
```

---

## Production-Ready State Persistence

For a real deployment (not the 2-hour challenge), the frontend would persist state:

```javascript
// On every state change, save to localStorage
useEffect(() => {
  if (gameState.gameId) {
    localStorage.setItem('wordgame', JSON.stringify(gameState));
  }
}, [gameState]);

// On mount, restore from localStorage
useEffect(() => {
  const saved = localStorage.getItem('wordgame');
  if (saved) {
    setGameState(JSON.parse(saved));
  }
}, []);
```

This way a page refresh doesn't lose the game. The `gameId` survives, and the player can continue where they left off without needing the backend to store sessions. Clean, stateless, and simple — the API doesn't need to change at all.
