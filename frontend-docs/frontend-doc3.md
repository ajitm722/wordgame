# Frontend Doc #3: Complete File Reference & Architecture Guide

## Table of Contents

1. [Overview](#1-overview)
2. [Directory Tree](#2-directory-tree)
3. [File Breakdown: Entry & Templates](#3-file-breakdown-entry--templates)
4. [File Breakdown: Context (Tier 1 — Global State)](#4-file-breakdown-context-tier-1--global-state)
5. [File Breakdown: Context (Tier 2 — Server Cache)](#5-file-breakdown-context-tier-2--server-cache)
6. [File Breakdown: HTTP Layer](#6-file-breakdown-http-layer)
7. [File Breakdown: Hooks (Tier 2 + Tier 3 — Game Logic)](#7-file-breakdown-hooks-tier-2--tier-3--game-logic)
8. [File Breakdown: Router](#8-file-breakdown-router)
9. [File Breakdown: Component Tree](#9-file-breakdown-component-tree)
10. [File Breakdown: Interfaces](#10-file-breakdown-interfaces)
11. [File Breakdown: Styles](#11-file-breakdown-styles)
12. [File Breakdown: Test Infrastructure](#12-file-breakdown-test-infrastructure)
13. [Config Files at Repo Root](#13-config-files-at-repo-root)
14. [How It All Connects: Data Flow](#14-how-it-all-connects-data-flow)
15. [Error Recovery Flow](#15-error-recovery-flow)
16. [Theme System](#16-theme-system)
17. [Testing Strategy](#17-testing-strategy)

---

## 1. Overview

This is a complete reference for the word game frontend — 33 source files, 5 config files, and 11 tests. Every file is listed, explained, and connected to the rest of the system.

### The Tech Stack (One Line Each)

| Library | Purpose |
|---|---|
| **React 18.3** | UI framework — components as functions |
| **TypeScript** | Type-safe JavaScript |
| **React Router v6** | Client-side URL routing |
| **TanStack Query v5** | Server state (mutation lifecycle) |
| **Axios** | HTTP client |
| **Webpack 5** | Bundler (TSX + SCSS → JS + CSS) |
| **Sass (SCSS)** | CSS with nesting + variables |
| **Jest 29** | Test runner |
| **React Testing Library** | Component rendering in tests |

### The Three Tiers

| Tier | Mechanism | What It Holds | File(s) |
|---|---|---|---|
| 1 — Global State | React Context + useReducer | Theme, apiBaseUrl config | `context/app.tsx` |
| 2 — Server State | TanStack Query (useQuery + useMutation) | Game data via cache, API call lifecycle (pending/error) | `context/query.tsx`, `hooks/useGame.ts` |
| 3 — Local State | React useState | gameId, guessedLetters | `hooks/useGame.ts` |

Game data is Tier 2 — read from TanStack Query's cache via `useQuery` (keyed on `gameId`) and updated via `setQueryData` in mutation `onSuccess` callbacks. Local state only holds `gameId` and `guessedLetters`.

---

## 2. Directory Tree

```
frontend/
├── index.tsx                          ← Entry point
├── index.scss                          ← Master stylesheet
├── templates/
│   └── react.ejs                      ← HTML shell template
│
├── components/
│   └── App/
│       ├── App.tsx                    ← Root component (wires providers + ThemeSync)
│       ├── App.test.tsx               ← Smoke test
│       └── index.ts                   ← Barrel export
│
├── context/
│   ├── app.tsx                        ← Tier 1: AppProvider + useReducer
│   └── query.tsx                      ← Tier 2: QueryClientProvider
│
├── hooks/
│   └── useGame.ts                     ← Game state hook (Tier 2 + Tier 3)
│
├── interfaces/
│   └── game.ts                        ← API response type definitions
│
├── layouts/
│   ├── CoreLayout/
│   │   ├── CoreLayout.tsx             ← App chrome (SiteTopNav + Outlet)
│   │   ├── CoreLayout.test.tsx        ← Smoke test
│   │   ├── _styles.scss               ← App chrome styles (theme-aware)
│   │   └── index.ts                   ← Barrel export
│   │
│   └── SiteTopNav/
│       ├── SiteTopNav.tsx             ← Nav bar (title + theme toggle)
│       ├── _styles.scss               ← Yellow gradient + toggle button
│       └── index.ts                   ← Barrel export
│
├── pages/
│   └── GamePage/
│       ├── GamePage.tsx               ← Main game screen (5 states)
│       ├── GamePage.test.tsx           ← 7 tests covering all states
│       ├── _styles.scss               ← Game styles (theme-aware)
│       └── index.ts                   ← Barrel export
│
├── router/
│   ├── index.tsx                      ← createBrowserRouter definition
│   └── paths.ts                       ← URL path constants
│
├── services/
│   └── entities/
│       └── game.ts                    ← API entity (newGame, guess)
│
├── styles/
│   └── var/
│       └── colors.scss                ← Color palette + theme CSS variables
│
├── test/
│   ├── test-setup.ts                  ← MSW lifecycle + Web API polyfills
│   ├── test-utils.tsx                 ← Custom render with providers
│   ├── mock-server.ts                 ← MSW setupServer (default export)
│   └── default-handlers.ts            ← Default route handlers
│
└── utilities/
    ├── endpoints.ts                   ← API path constants
    └── sendRequest.ts                 ← Single HTTP choke point
```

### File Count: 33 source files

---

## 3. File Breakdown: Entry & Templates

### `frontend/index.tsx` (Entry Point)

**Purpose:** The Webpack entry point. Mounts the React app into the DOM.

```tsx
import { createRoot } from "react-dom/client";
import { App } from "components/App";
import "index.scss";

createRoot(document.getElementById("app")!).render(<App />);
```

**What happens:**

1. Imports `App` (the root React component)
2. Imports `index.scss` — Webpack sees this and bundles all SCSS dependencies
3. Calls `createRoot(...).render(<App />)` — React takes over the `<div id="app">` in the HTML shell

**The `!` (non-null assertion):** Tells TypeScript "trust me, this element exists." The HTML template guarantees `<div id="app">` is present.

---

### `frontend/templates/react.ejs` (HTML Shell)

**Purpose:** The HTML template that Webpack's HtmlWebpackPlugin uses. This is the only HTML file the browser loads.

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Word Game</title>
</head>
<body>
  <div id="app"></div>
</body>
</html>
```

**What HtmlWebpackPlugin does:** Injects `<script>` and `<link>` tags for the bundled JS and CSS into this template. The final output in `assets/index.html` will have the bundle paths auto-inserted, e.g.:
```html
<script src="/assets/bundle.5071ee47.js"></script>
<link href="/assets/bundle.c8adf365.css" rel="stylesheet" />
```

---

### `frontend/index.scss` (Master Stylesheet)

**Purpose:** Pulls in every component's co-located `_styles.scss` via `@use`. This is the single SCSS entry point.

```scss
@use "styles/var/colors";          // Color palette + theme variables
@use "layouts/CoreLayout/styles" as layout;
@use "layouts/SiteTopNav/styles" as nav;
@use "pages/GamePage/styles" as game;
```

**The `@use` rule:** Like Go imports. Each `@use "path/styles"` tells Sass to compile `path/_styles.scss` and include it. The `as` namespace prevents name collisions. Every component's styles end up in the final `bundle.css`.

**Why this exists:** Instead of one giant CSS file, each component has its own `_styles.scss` that only styles that component. The master sheet imports them all. If you delete a component, you delete its `_styles.scss` and the import — no dead CSS.

---

## 4. File Breakdown: Context (Tier 1 — Global State)

### `context/app.tsx` (AppProvider + useReducer)

**Purpose:** Global state available to every component without prop-drilling. Holds theme preference and config.

```tsx
interface IAppState {
  theme: "light" | "dark";
  config: {
    apiBaseUrl: string;     // "http://localhost:1337"
  };
}
```

**Key design decisions:**

- **Single context** (not split into state + dispatch like Fleet). Simpler to consume at the cost of broader re-renders. For an app this size, the perf difference is negligible.
- **Single action type** (`SET_THEME`). Config is set at startup and never changes.
- **PascalCase interface name** with `I` prefix: `IAppState`, `IAppContext`.

**How consumers use it:**

```tsx
import { useAppContext } from "context/app";

function MyComponent() {
  const { state, dispatch } = useAppContext();
  // state.theme → "light" | "dark"
  // state.config.apiBaseUrl → "http://localhost:1337"
  // dispatch({ type: "SET_THEME", payload: "dark" }) → switches theme
}
```

**The null guard:** If a component calls `useAppContext()` outside of `<AppProvider>`, it throws a clear error at development time.

---

### `context/query.tsx` (QueryProvider)

**Purpose:** Wraps the app with TanStack Query's `QueryClientProvider` so any component can call `useMutation` to make API calls and track their lifecycle.

```tsx
const FIVE_MINUTES = 5 * 60 * 1000;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: FIVE_MINUTES,          // 5 min before refetch
      retry: 1,                          // retry once on failure
      refetchOnWindowFocus: false,       // don't refetch on tab switch
    },
  },
});
```

**`FIVE_MINUTES = 5 * 60 * 1000`:** JavaScript's `Date.now()` and `setTimeout()` measure time in **milliseconds** (1/1000th of a second), so "5 minutes" has to be written as the math to convert minutes → milliseconds: `5 × 60 sec × 1000 ms = 300,000 ms`. Extracting it as a named constant is just for readability — you'll see this pattern everywhere a duration is passed to a JS API.

**`defaultOptions`** configures how every query behaves by default. None of these three affect `useMutation` directly — they only configure `useQuery` reads. We don't use `useQuery` in this app (see below), so the values are sensible defaults for future use.

| Setting | Effect | Value |
|---|---|---|
| `staleTime` | How long cached data is considered "fresh" | `FIVE_MINUTES` (5 min) |
| `retry` | How many times to retry a failed query | `1` (retry once) |
| `refetchOnWindowFocus` | Re-fetch when the user tabs back to the page | `false` |

#### What we use vs. what we don't

TanStack Query has two hooks:

- **`useQuery`** — "get me this data" (reads, with caching + auto-refetch)
- **`useMutation`** — "send this write" (posts, with pending/error state)

We use both hooks: `useQuery` to read game data from the cache, and `useMutation` to write to the cache and the server. From `useGame.ts`:

```ts
const { data: gameData } = useQuery<IGuessResponse>({
  queryKey: ["game", gameId],
  queryFn: () => gameAPI.getGame(gameId!),
  enabled: !!gameId,
});

const newGameMutation = useMutation({
  mutationFn: gameAPI.newGame,                                  // POST /api/v1/new
  onSuccess: (data) => { queryClient.setQueryData(["game", data.id], data); setGameId(data.id); },
  onError:    ()     => { setGameId(null); },
});
```

The `queryClient` tracks four lifecycle fields per mutation: `isPending`, `isError`, `error`, `data`. `GamePage` reads `isPending` / `isError` to decide what to render.

We don't use `useQuery` for game data because the data is **owned by one component** (just `GamePage`), changes **synchronously on user action**, and a disabled-query pattern (`enabled: false`) doesn't reliably observe cache writes. Game data lives in plain `useState` in `useGame.ts` — the `queryClient` is here only because `useMutation` needs it to function.

---

## 5. File Breakdown: Context (Tier 2 — Server Cache)

(The `context/query.tsx` file already covers Tier 2 setup. See Section 7 for how `useGame.ts` uses mutations.)

---

## 6. File Breakdown: HTTP Layer

### `utilities/endpoints.ts` (API Path Constants)

**Purpose:** Single source of truth for API URL paths. No hardcoded strings anywhere else.

```ts
const API_BASE = "/api/v1";
export default {
  NEW_GAME: `${API_BASE}/new`,
  GUESS: `${API_BASE}/guess`,
};
```

**Convention:** All paths start with `/api/v1/`. The Go mux uses a subrouter at this prefix. The browser makes requests to the same origin (no CORS needed).

---

### `utilities/sendRequest.ts` (HTTP Choke Point)

**Purpose:** Every API call in the app goes through this single function.

```ts
export async function sendRequest<T>(method: "GET" | "POST", path: string, data?: unknown): Promise<T>
```

**Type parameter `T`:** Callers specify the expected response type:

```ts
// Returns Promise<INewGameResponse>
sendRequest<INewGameResponse>("POST", endpoints.NEW_GAME);

// Returns Promise<IGuessResponse>
sendRequest<IGuessResponse>("POST", endpoints.GUESS, { id, guess: "A" });
```

**Error handling:**

```ts
catch (err) {
  const axiosErr = err as AxiosError<{ message?: string }>;
  throw new Error(
    axiosErr.response?.data?.message ||  // try server's error message
    axiosErr.message ||                   // try Axios message (e.g. "Network Error")
    "Network error"                       // fallback
  );
}
```

All errors become `Error` objects with a string message, regardless of whether it was a server 500, a network timeout, or a DNS failure.

---

### `services/entities/game.ts` (Game API Entity)

**Purpose:** Thin wrapper around `sendRequest` typed for game API calls.

```ts
export default {
  newGame: () => sendRequest<INewGameResponse>("POST", endpoints.NEW_GAME),
  guess: (id: string, guess: string) =>
    sendRequest<IGuessResponse>("POST", endpoints.GUESS, { id, guess }),
};
```

**Why an entity service pattern:** The `useGame` hook doesn't call `sendRequest` directly. It calls `gameAPI.newGame()` or `gameAPI.guess(id, guess)`. This layer:
- Is the **only** file that knows the concrete API endpoints
- Can be **mocked in tests** via MSW handlers in `test/default-handlers.ts`
- Can be **extended** with more game-related API calls without touching other files

---

## 7. File Breakdown: Hooks (Tier 2 + Tier 3 — Game Logic)

### `hooks/useGame.ts` (The Game Hook)

**Purpose:** The heart of the game logic. Manages all game state and exposes a clean API for `GamePage` to render.

**Imports:**

```ts
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, useCallback } from "react";
import gameAPI from "services/entities/game";
import type { IGuessResponse } from "interfaces/game";
```

**Two `useState` variables (Tier 3):**

```ts
const [gameId, setGameId] = useState<string | null>(null);
const [guessedLetters, setGuessedLetters] = useState<string[]>([]);
```

- `gameId`: The active game's ID from the server. `null` means "no game in progress" (idle state). Stored locally — it drives the query key for `useQuery` and the body of the next `guess` mutation.
- `guessedLetters`: Tracked locally because the API doesn't know which letters you've guessed. Set optimistically (before the API call) so the keyboard can disable the button immediately.

**One `useQuery` (Tier 2 — game data via cache):**

```ts
const { data: gameData } = useQuery<IGuessResponse>({
  queryKey: ["game", gameId],
  queryFn: () => gameAPI.getGame(gameId!),
  enabled: !!gameId,
});
```

- `queryKey: ["game", gameId]` — cache key includes the game ID. A new game means a new cache entry; a missing `gameId` means no query runs.
- `enabled: !!gameId` — the query only fires when there's an active game. On idle (`gameId === null`), no `getGame` request is made.
- `useQuery` reads from the cache. The cache is populated by `setQueryData` calls in the mutations below, so the request to `getGame` only fires on a cold start (page refresh with a `gameId` in scope). In normal flow, mutations write to the cache and `useQuery` re-renders with the new data — no extra network request.

**Two `useMutation` hooks (Tier 2 — write lifecycle):**

```ts
const newGameMutation = useMutation({
  mutationFn: gameAPI.newGame,
  onSuccess: (data) => {
    queryClient.setQueryData(["game", data.id], data); // seed cache
    setGameId(data.id);                                // activate game
    setGuessedLetters([]);                             // clear previous guesses
  },
  onError: () => {
    setGameId(null);                                   // back to idle
    setGuessedLetters([]);
  },
});

const guessMutation = useMutation({
  mutationFn: ({ id, guess }: { id: string; guess: string }) =>
    gameAPI.guess(id, guess),
  onSuccess: (data, variables) => {
    queryClient.setQueryData(["game", variables.id], data); // update cache
  },
  onError: () => {
    setGameId(null);                                   // game lost — back to idle
    setGuessedLetters([]);
  },
});
```

- **Cache-first pattern:** `onSuccess` calls `queryClient.setQueryData(["game", id], data)` to write the new server response into the cache. The `useQuery` above observes this cache change and re-renders. No extra `GET` is needed after a `POST`.
- `guessMutation.onSuccess` uses `variables.id` (the `id` passed into the mutation) so the cache write targets the right entry.
- Both mutations clear `gameId` on error — this resets the component to the idle state, since `useQuery`'s `enabled: false` will skip the next `getGame` call.

**Derived values:**

```ts
const current = gameData?.current ?? null;
const guessesRemaining = gameData?.guesses_remaining ?? null;
const word = gameData?.word ?? null;

const isWon = current !== null && !current.includes("_");
const isLost = current !== null && guessesRemaining !== null && guessesRemaining <= 0;
```

- `current`: The word display (e.g. `"_PP__"`)
- `guessesRemaining`: Counter for the UI
- `word`: Secret word — only present in the response when the game ends (won or lost). `useGame` always exposes it; `GamePage` chooses whether to render it.
- `isWon`: All letters revealed (no underscores in `current`)
- `isLost`: No guesses remaining

**The `newGame` and `makeGuess` callbacks:**

```ts
const newGame = useCallback(() => {
  newGameMutation.mutate();
}, [newGameMutation]);

const makeGuess = useCallback(
  async (letter: string) => {
    if (!gameId) return;                                // no active game
    setGuessedLetters((prev) => [...prev, letter]);     // optimistic update
    await guessMutation.mutateAsync({ id: gameId, guess: letter });
  },
  [gameId, guessMutation]
);
```

- `newGame`: thin wrapper around `newGameMutation.mutate()` so `GamePage` doesn't need to know about the mutation directly.
- `makeGuess`: early-returns if no `gameId` (so click handlers on the keyboard are safe in idle state). Then optimistically adds the letter to `guessedLetters` (so the button disables immediately) and awaits the mutation.
- `gameId` in deps: ensures the callback captures the latest game ID. Without it, every guess would send `{ id: null }` after a new game starts.

**Return value — the hook's public API:**

```ts
return {
  gameId, current, word, guessesRemaining, guessedLetters,
  newGame, makeGuess,
  isWon, isLost,
  isLoading: newGameMutation.isPending,
  isPendingGuess: guessMutation.isPending,
  isError: newGameMutation.isError || guessMutation.isError,
  error: newGameMutation.error || guessMutation.error,
};
```

- `isLoading` tracks only `newGameMutation` (starting a game). The button shows "Starting..." while this is true.
- `isPendingGuess` tracks only `guessMutation` (keyboard disables while this is true).
- `isError` and `error` combine both mutations — a single error overlay handles either failure.

---

## 8. File Breakdown: Router

### `router/paths.ts` (URL Constants)

```ts
export default {
  HOME: "/",
};
```

Only one route for now. All URL strings live here — no hardcoded paths elsewhere.

---

### `router/index.tsx` (Route Tree)

```tsx
const router = createBrowserRouter([
  {
    path: PATHS.HOME,
    element: <CoreLayout />,
    children: [{ index: true, element: <GamePage /> }],
  },
]);

export function AppRouter() {
  return <RouterProvider router={router} />;
}
```

**Structure:**

- `CoreLayout` is the root layout — it always renders `SiteTopNav` at the top
- `GamePage` is the index (default) child route rendered inside `CoreLayout`'s `<Outlet />`
- Future pages (e.g. `/settings`) would be added as more child routes

**`createBrowserRouter` vs `BrowserRouter`:** `createBrowserRouter` uses the newer data router API (React Router v6.4+). It doesn't use a `<Routes>` wrapper — instead it uses a declarative route config object. `RouterProvider` feeds the router to React.

---

## 9. File Breakdown: Component Tree

### `components/App/App.tsx` (Root Component)

```
AppProvider
  └── ThemeSync (null-rendering, syncs theme to <html>)
      └── QueryProvider
          └── AppRouter
```

**`ThemeSync`:** Renders nothing (`return null`). Its `useEffect` runs after each render:

```tsx
useEffect(() => {
  document.documentElement.dataset.theme = state.theme;
}, [state.theme]);
```

This sets `<html data-theme="dark">` or `<html data-theme="light">`. The SCSS in `colors.scss` uses `[data-theme="dark"]` to apply dark-mode CSS variables.

**Provider order matters:** `AppProvider` wraps everything (Tier 1 context must be available to all). Then `ThemeSync` reads from it. Then `QueryProvider` wraps the router (Tier 2 must be available to route components).

---

### `layouts/CoreLayout/CoreLayout.tsx` (App Chrome)

```tsx
export function CoreLayout() {
  return (
    <div className="core-layout">
      <SiteTopNav />
      <main className="core-layout__content">
        <Outlet />   ← child route content rendered here
      </main>
    </div>
  );
}
```

**`<Outlet />`:** React Router component. Renders the active child route (currently `GamePage`). Future routes would also render here, keeping the nav bar consistent.

---

### `layouts/SiteTopNav/SiteTopNav.tsx` (Navigation Bar)

**Purpose:** Title + theme toggle button.

```tsx
export function SiteTopNav() {
  const { state, dispatch } = useAppContext();

  function toggleTheme() {
    dispatch({
      type: "SET_THEME",
      payload: state.theme === "light" ? "dark" : "light",
    });
  }

  return (
    <nav className="site-top-nav">
      <span className="site-top-nav__title">Word Game</span>
      <button className="site-top-nav__theme-btn" onClick={toggleTheme}>
        {state.theme === "light" ? "🌙 Dark" : "☀️ Light"}
      </button>
    </nav>
  );
}
```

**Flow:** Click button → `dispatch({ type: "SET_THEME", payload: "dark" })` → `appReducer` updates state → `AppProvider` re-renders → `ThemeSync.useEffect` sets `document.documentElement.dataset.theme = "dark"` → CSS variables switch → every component using `var(--color-*)` instantly updates.

---

### `pages/GamePage/GamePage.tsx` (Main Game Screen)

**Five states rendered via early returns:**

```
1. isError  →  "Error: {message}" + "Try Again" button
2. current === null  →  "Word Game" title + "New Game" button
3. isWon  →  "You Won!" + revealed tiles + "Play Again" button
4. isLost  →  "Game Over" + secret word + "Try Again" button
5. default (playing)  →  board tiles + keyboard + guesses counter + "New Game" button
```

**State priority:** Error trumps idle trumps won/lost trumps playing. Because `isError` is checked first, an API failure during a game will show the error overlay.

**Playing state rendering:**

- Board: `current.split("").map(...)` renders each character in a tile. Underscores get an empty tile style; revealed letters get `tile--revealed` (green).
- Keyboard: 26 buttons, A-Z. Each button is disabled if `guessedLetters.includes(letter)` or `isPendingGuess` (prevents double-clicks).
- "Guesses remaining" counter: shows `guessesRemaining` from the API.
- "New Game" button: always accessible during play.

---

## 10. File Breakdown: Interfaces

### `interfaces/game.ts` (API Response Types)

```ts
export interface INewGameResponse {
  id: string;
  current: string;           // e.g. "_PP__"
  guesses_remaining: number; // e.g. 6
}

export interface IGuessResponse {
  id: string;
  current: string;
  guesses_remaining: number;
  word?: string;             // only present when game ends (won or lost)
}

export interface IGameError {
  message: string;
  code?: string;
}
```

**Why two types?** `IGuessResponse` has an optional `word` field that `INewGameResponse` doesn't. The `useGame` hook unifies the cache as `IGuessResponse` (since `useQuery` is typed to one shape) and reads `gameData?.word` directly — no cast needed.

---

## 11. File Breakdown: Styles

### `styles/var/colors.scss` (Theme Variables)

**Named palette variables:**

```scss
$color-nav-yellow-start: #facc15;
$color-nav-yellow-end:   #f59e0b;
$color-primary:          #1976d2;
$color-success:          #4caf50;
$color-error:            #d32f2f;
```

**CSS custom properties (theme-aware):**

```scss
[data-theme="light"] {
  --color-text:        #333;
  --color-bg:          #fff;
  --color-tile-border: #ccc;
  --color-key-bg:      #f5f5f5;
  // ... more vars
}

[data-theme="dark"] {
  --color-text:        #e0e0e0;
  --color-bg:          #121212;
  --color-tile-border: #555;
  --color-key-bg:      #333;
  // ... more vars
}
```

**How theme switching works:** `data-theme` attribute is set on `<html>` by `ThemeSync`. CSS browsers apply the matching rule set. All components use `var(--color-text)`, `var(--color-bg)`, etc. No component needs to know about light/dark — they just reference the variables.

---

### `layouts/CoreLayout/_styles.scss`

```scss
.core-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
  color: var(--color-text);
  transition: background 0.3s, color 0.3s;

  &__content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
```

The smooth `transition` on background and color makes theme changes feel fluid.

---

### `layouts/SiteTopNav/_styles.scss`

```scss
.site-top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 2rem;
  background: linear-gradient(135deg, #facc15, #f59e0b);
  color: #1a1a2e;

  &__theme-btn {
    background: rgba(0, 0, 0, 0.15);
    border: 1px solid rgba(0, 0, 0, 0.25);
    border-radius: 6px;
    padding: 0.35rem 0.75rem;
    cursor: pointer;
    color: #1a1a2e;
    font-weight: 600;
  }
}
```

Yellow gradient: `#facc15` → `#f59e0b` (warm yellow to amber). Dark text (`#1a1a2e`) for contrast.

---

### `pages/GamePage/_styles.scss`

All game UI styles using `var(--color-*)` for theme awareness:

- **Tiles**: `var(--color-tile-bg)`, `var(--color-tile-border)` — reveal state uses hardcoded green (`#4caf50`, `#e8f5e9`)
- **Keyboard keys**: `var(--color-key-bg)`, `var(--color-key-border)`, `var(--color-text)`
- **Buttons**: `var(--color-btn-bg)`, `var(--color-btn-border)`, `var(--color-text)`
- **Primary button**: hardcoded blue (`#1976d2`) — it's a call-to-action that should always look the same

---

## 12. File Breakdown: Test Infrastructure

### `test/test-setup.ts` (MSW Lifecycle + Web API Polyfills)

Loaded via Jest's `setupFilesAfterEnv` (runs after the test framework is installed, so `beforeAll`/`afterEach`/`afterAll` are available).

```ts
import "@testing-library/jest-dom";

const { WritableStream } = require("node:stream/web");
const { MessagePort, MessageChannel } = require("node:worker_threads");

(globalThis as any).WritableStream = WritableStream;
(globalThis as any).MessagePort = MessagePort;
(globalThis as any).MessageChannel = MessageChannel;
(globalThis as any).Event = Event;
(globalThis as any).EventTarget = EventTarget;

const mockServer = require("./mock-server").default;

beforeAll(() => mockServer.listen());
afterEach(() => mockServer.resetHandlers());
afterAll(() => mockServer.close());
```

**Web API polyfills:** `jest-fixed-jsdom@0.0.8` exposes `Request`, `Response`, `fetch`, `TransformStream`, `ReadableStream`, `TextEncoder`/`TextDecoder`, `Headers`, `FormData`, `Blob` from Node's globals — but not `WritableStream`, `MessagePort`, `MessageChannel`, `Event`, `EventTarget`. We polyfill the missing five so MSW's source files can load. Polyfills happen BEFORE the `mockServer` import (uses `require` instead of `import` to control order) so MSW sees the globals at module-load time.

---

### `test/test-utils.tsx` (Custom Render)

**Purpose:** Wraps components in the same providers the real app uses, so tests don't duplicate provider setup.

```tsx
function customRender(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  function AllProviders({ children }) {
    return (
      <AppProvider>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </AppProvider>
    );
  }

  return render(ui, { wrapper: AllProviders });
}
```

**Fresh `QueryClient` per render:** Prevents cache bleed between tests. Each test starts with a clean slate.

---

### `test/mock-server.ts` (MSW setupServer)

A one-liner that exports the MSW Node server as a default export. Tests import it to register overrides via `mockServer.use(...)`.

```ts
import { setupServer } from "msw/node";
import handlers from "./default-handlers";

const mockServer = setupServer(...handlers);

export default mockServer;
```

### `test/default-handlers.ts` (Default Route Handlers)

Default route handlers that MSW serves when no per-test override is registered. The defaults cover the three game endpoints with sensible fallbacks:

```ts
import { http, HttpResponse } from "msw";

export const handlers = [
  http.post("/api/v1/new", () => {
    return HttpResponse.json({
      id: "test-id", current: "_____", guesses_remaining: 6,
    });
  }),
  http.get("/api/v1/game/:id", ({ params }) => {
    return HttpResponse.json({
      id: params.id, current: "_____", guesses_remaining: 6,
    });
  }),
  http.post("/api/v1/guess", () => {
    return HttpResponse.json({
      id: "test-id", current: "_____", guesses_remaining: 5,
    });
  }),
];

export default handlers;
```

`onUnhandledRequest: "error"` is set in `mockServer.listen()` (via `test-setup.ts`) so any request without a matching handler fails the test loudly — catching accidental unmocked calls.

---

### `components/App/App.tests.tsx` (Root Component Test)

```tsx
test("renders idle game page with New Game button", () => {
  render(<App />);
  expect(screen.getByRole("button", { name: /new game/i })).toBeInTheDocument();
});
```

Smoke test — verifies the entire component tree mounts without crashing and shows the idle state.

---

### `layouts/CoreLayout/CoreLayout.tests.tsx` (Layout Test)

```tsx
test("renders SiteTopNav", () => {
  render(<CoreLayout />);
  expect(screen.getByText("Word Game")).toBeInTheDocument();
});
```

Verifies the nav bar renders. The `<Outlet />` renders nothing (no router in this test), but the nav bar always shows.

---

### `pages/GamePage/GamePage.tests.tsx` (7 Tests)

The main test file. Uses MSW (`mockServer.use()`) to override the default handlers per test.

```ts
import { http, HttpResponse } from "msw";
import mockServer from "test/mock-server";

test("renders won state", async () => {
  mockServer.use(
    http.post("/api/v1/guess", () => {
      return HttpResponse.json({
        id: "test-id",
        current: "APPLE",
        guesses_remaining: 6,
      });
    })
  );
  // ... render + click + assert
});
```

**Tests (in order):**

| # | Test | What it verifies |
|---|---|---|
| 1 | `renders idle state with New Game button` | Initial state — title + button |
| 2 | `starts a game when New Game is clicked` | Click → API call → board renders |
| 3 | `renders keyboard in playing state` | All 26 letters visible |
| 4 | `disables guessed letter and updates remaining guesses` | Click "Z" → button disabled + guesses decrement |
| 5 | `renders won state` | No underscores → "You Won!" + "Play Again" |
| 6 | `renders lost state with secret word` | 0 guesses → "Game Over" + word shown |
| 7 | `renders error state with retry button` | API rejects → error message + "Try Again" |

**Why MSW instead of `jest.mock()`:** We follow Fleet's testing pattern (see `Fleet_Frontend_Deep_Dive.md` §14.3). MSW intercepts at the network level, so tests verify the real request/response contract rather than mocking at the module level.

---

## 13. Config Files at Repo Root

### `package.json`

```json
{
  "scripts": {
    "build": "webpack --mode production",
    "test": "jest --config jest.config.js"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.60.0",
    "axios": "^1.7.0",
    "react": "18.3.1",
    "react-dom": "18.2.0",
    "react-router-dom": "^6.26.0"
  },
  "devDependencies": {
    "@testing-library/jest-dom": "^6.6.0",
    "@testing-library/react": "^14.3.0",
    "@testing-library/user-event": "^14.5.0",
    "@types/jest": "^29.5.0",
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "autoprefixer": "^10.4.0",
    "css-loader": "^6.11.0",
    "fork-ts-checker-webpack-plugin": "^9.0.0",
    "html-webpack-plugin": "^5.6.0",
    "identity-obj-proxy": "^3.0.0",
    "jest": "^29.7.0",
    "jest-environment-jsdom": "^29.7.0",
    "mini-css-extract-plugin": "^2.9.0",
    "msw": "^2.6.0",
    "postcss": "^8.4.0",
    "postcss-loader": "^8.1.0",
    "sass": "^1.80.0",
    "sass-loader": "^16.0.0",
    "ts-jest": "^29.2.0",
    "ts-loader": "^9.5.0",
    "typescript": "^5.6.0",
    "webpack": "^5.96.0",
    "webpack-cli": "^5.1.0"
  }
}
```

**No `start` script** — the old `webpack serve` command was removed. The dev workflow uses `make dev` (Go server + webpack --watch).

---

### `tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "baseUrl": ".",
    "paths": {
      "*": ["./frontend/*"]     ← bare imports: "hooks/useGame" resolves to "./frontend/hooks/useGame"
    }
  },
  "include": ["frontend/**/*"]
}
```

**Key settings:**
- `jsx: "react-jsx"` — the modern JSX transform (no `import React` needed in components)
- `moduleResolution: "bundler"` — Node.js-style resolution with ESM support
- `paths: { "*": ["./frontend/*"] }` — enables bare imports like `import { useGame } from "hooks/useGame"`

---

### `webpack.config.js`

```js
module.exports = (env, argv) => {
  const isProd = argv.mode === "production";
  return {
    entry: "./frontend/index.tsx",
    output: {
      path: "./assets",
      filename: isProd ? "bundle.[contenthash:8].js" : "bundle.js",
      publicPath: "/assets/",
      clean: true,
    },
    resolve: {
      extensions: [".ts", ".tsx", ".js", ".jsx"],
      modules: ["./frontend", "node_modules"],  ← bare imports
    },
    module: {
      rules: [
        { test: /\.tsx?$/, use: "ts-loader" },
        { test: /\.scss$/, use: [MiniCssExtractPlugin, "css-loader", "postcss-loader", "sass-loader"] },
      ],
    },
    plugins: [
      new HtmlWebpackPlugin({ template: "./frontend/templates/react.ejs" }),
      new MiniCssExtractPlugin({ filename: "bundle.[contenthash:8].css" }),
      new ForkTsCheckerWebpackPlugin({ typescript: { configFile: "tsconfig.json" } }),
    ],
    performance: {
      maxAssetSize: 400000,      ← raised from 244 KiB default
      maxEntrypointSize: 400000,
    },
  };
};
```

**Note:** The `performance` block was added to suppress warnings about bundle size. 271 KiB is normal for a React app — React DOM alone is 131 KiB. The limit was raised to 400 KiB so the warning is still meaningful (fires if a large library is accidentally included).

---

### `jest.config.js`

```js
module.exports = {
  rootDir: "frontend",
  testEnvironment: "jsdom",
  transform: { "^.+\\.tsx?$": ["ts-jest", { diagnostics: false }] },
  moduleNameMapper: { "\\.(css|scss)$": "identity-obj-proxy" },
  moduleDirectories: ["<rootDir>", "node_modules"],
};
```

**Key settings:**
- `rootDir: "frontend"` — all paths relative to `frontend/`
- `transform: { "^.+\\.tsx?$": "ts-jest" }` — compiles TypeScript in tests
- `moduleNameMapper: { "\\.(css|scss)$": "identity-obj-proxy" }` — SCSS imports return proxy objects instead of real CSS
- `moduleDirectories` — enables bare imports in tests (matching webpack's `resolve.modules`)

---

### `postcss.config.js`

```js
module.exports = {
  plugins: [require("autoprefixer")],
};
```

Adds vendor prefixes to CSS (e.g. `-webkit-`, `-moz-`) for older browser support.

---

## 14. How It All Connects: Data Flow

### Starting a New Game

```
User clicks "New Game" button
    │
    ▼
GamePage: onClick={newGame}
    │  newGame is from useGame()
    ▼
useGame.newGame: newGameMutation.mutate()
    │
    ▼
useMutation.mutationFn → gameAPI.newGame()
    │
    ▼
gameAPI.newGame → sendRequest("POST", endpoints.NEW_GAME)
    │
    ▼
Axios POST /api/v1/new → Go server → { id, current, guesses_remaining }
    │
    ▼
sendRequest returns Promise<INewGameResponse>
    │
    ▼
useMutation.onSuccess(data)
    │  queryClient.setQueryData(["game", data.id], data)  ← seed cache
    │  setGameId(data.id)                                  ← activate game
    │  setGuessedLetters([])                               ← clear previous state
    ▼
useQuery re-renders (reads from cache)
    │  current !== null → playing state
    │  Board shows underscore tiles
    │  Keyboard (A-Z) rendered
    ▼
User sees the game board
```

### Making a Guess

```
User clicks keyboard letter "P"
    │
    ▼
GamePage: onClick={() => makeGuess("P")}
    │  isPendingGuess || guessedLetters.includes("P") → disabled early return
    │  (button already disabled if guessed)
    ▼
useGame.makeGuess("P"):
  1. setGuessedLetters(prev => [...prev, "P"])  ← OPTIMISTIC
  2. guessMutation.mutateAsync({ id: gameId, guess: "P" })
    │
    ▼
useMutation.mutationFn → gameAPI.guess(id, "P")
    │
    ▼
sendRequest("POST", endpoints.GUESS, { id, guess: "P" })
    │
    ▼
Axios POST /api/v1/guess → Go server → { id, current, guesses_remaining, word? }
    │
    ├── Success → onSuccess(data, variables):
    │     queryClient.setQueryData(["game", variables.id], data)  ← update cache
    │     useQuery re-renders with new current value
    │     If current has no "_" → isWon state → "You Won!"
    │     If guesses_remaining is 0 → isLost state → "Game Over"
    │
    └── Error → onError(): setGameId(null), setGuessedLetters([])
          useQuery disabled (no gameId) → current becomes null
          GamePage re-renders to idle state
          User clicks "New Game" to try again
```

---

## 15. Error Recovery Flow

### Scenario: Server Restarts Mid-Game

```
1. User starts a game → queryClient cache has ["game", id] entry
2. Server restarts (loses in-memory game state)
3. User makes a guess "Z"

   makeGuess("Z"):
     setGuessedLetters(["Z"])            ← optimistic
     guessMutation.mutateAsync(...)      ← POST /api/v1/guess with old game ID

4. Go server returns 404 (game not found)

5. guessMutation.onError fires:
     setGameId(null)                     ← disables useQuery
     setGuessedLetters([])               ← clear local state

6. useQuery enabled: !!gameId → false → no getGame call
   gameData becomes undefined → current, guessesRemaining, word all become null
   GamePage re-renders:
     current → null                      ← idle state
     isError → false (mutation error cleared)
     Renders: "Word Game" + "New Game"

7. User clicks "New Game":
     newGameMutation.mutate()            ← fresh POST /api/v1/new
     onSuccess: setQueryData + setGameId ← new game starts
```

**Why this works:** Both `newGameMutation.onError` and `guessMutation.onError` call `setGameId(null)`. This flips `useQuery`'s `enabled` to `false`, so no `getGame` request fires for the dead game. The component re-renders to the idle state deterministically — no stale game ID lingers, no partial state remains.

---

## 16. Theme System

### Architecture

```
User clicks theme toggle in SiteTopNav
    │  dispatch({ type: "SET_THEME", payload: "dark" })
    ▼
useReducer: appReducer returns { ...state, theme: "dark" }
    │
    ▼
AppProvider re-renders children
    │
    ├── SiteTopNav reads state.theme → shows "☀️ Light" button text
    │
    └── ThemeSync reads state.theme
          useEffect → document.documentElement.dataset.theme = "dark"
            │
            ▼
            CSS: [data-theme="dark"] rules activate
              --color-bg: #121212
              --color-text: #e0e0e0
              --color-key-bg: #333
              ...
                │
                ▼
                CoreLayout background: var(--color-bg) → #121212
                GamePage tile border: var(--color-tile-border) → #555
                Keyboard key: var(--color-key-bg) → #333
                button: var(--color-btn-bg) → #333
                All themed elements update simultaneously
```

### Theme-Aware Files

| File | What it does |
|---|---|
| `context/app.tsx` | Holds `state.theme`, dispatches `SET_THEME` |
| `components/App/App.tsx` | `ThemeSync` writes `data-theme` to `<html>` |
| `layouts/SiteTopNav/SiteTopNav.tsx` | Toggle button reads/dispatches theme |
| `styles/var/colors.scss` | Defines both `[data-theme="light"]` and `[data-theme="dark"]` |
| `layouts/CoreLayout/_styles.scss` | Uses `var(--color-bg)`, `var(--color-text)` |
| `pages/GamePage/_styles.scss` | Uses `var(--color-*)` for tiles, keys, buttons |

### Adding a New Theme-Aware Component

1. Use `var(--color-*)` values for any property that should change with theme
2. If a new variable is needed, add it to both `[data-theme="light"]` and `[data-theme="dark"]` in `colors.scss`
3. No need to touch the theme logic — it's inherited from `CoreLayout`'s CSS variables

---

## 17. Testing Strategy

### Mocking Approach

We mock the API at the network level using MSW v2 (matches Fleet). Default handlers live in `test/default-handlers.ts`; per-test overrides use `mockServer.use()`:

```ts
// GamePage.tests.tsx
import { http, HttpResponse } from "msw";
import mockServer from "test/mock-server";

test("renders won state", async () => {
  mockServer.use(
    http.post("/api/v1/guess", () => {
      return HttpResponse.json({
        id: "test-id",
        current: "APPLE",
        guesses_remaining: 6,
      });
    })
  );
  // ... render + click + assert
});
```

**Why this works:** The `useGame` hook calls `gameAPI.guess(id, letter)` which calls `sendRequest("POST", "/api/v1/guess", ...)`. Axios sends the request, MSW intercepts at the network level and returns the mock response. The component doesn't know it's being tested — the API call works exactly like in production, just with canned responses.

### Provider Setup

Each test renders through `test-utils.tsx`'s custom `render()`:

```
AppProvider (Tier 1 context)
  └── QueryClientProvider (fresh client per render)
        └── <Component under test>
```

Fresh `QueryClient` per render prevents test pollution. `retry: false` ensures mutations don't retry (which would cause async timing issues in tests).

### Test Pattern

```ts
test("renders won state", async () => {
  // 1. Arrange: override MSW handler for this test
  mockServer.use(
    http.post("/api/v1/guess", () => {
      return HttpResponse.json({
        id: "test-id", current: "APPLE", guesses_remaining: 6,
      });
    })
  );

  // 2. Act: simulate user actions
  const user = userEvent.setup();
  renderPage();
  await user.click(screen.getByRole("button", { name: /new game/i }));
  await waitFor(() => {
    expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
  });
  await user.click(screen.getByRole("button", { name: "A" }));

  // 3. Assert: check the UI
  await waitFor(() => {
    expect(screen.getByText("You Won!")).toBeInTheDocument();
  });
});
```

**`waitFor`:** Retries the assertion until it passes or times out. Needed because state updates from mutation callbacks are async.

**No `act()` wrappers:** React Testing Library's `waitFor` and `userEvent` handle `act()` internally.

### What the 11 Tests Cover

| Test File | Tests | Coverage |
|---|---|---|
| `App.tests.tsx` | 1 | Root component mounts without crashing |
| `CoreLayout.tests.tsx` | 1 | Nav bar renders |
| `GamePage.tests.tsx` | 7 | All 5 states + keyboard + guess interaction |
| `SiteTopNav.tests.tsx` | 2 | Brand text + theme toggle button render |

Total: 11 tests, 4 suites, all passing.
