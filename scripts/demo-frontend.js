// Code-driven browser recording for the Word Game frontend.
// Mirrors the spirit of `demo.tape` (VHS) but for the React UI: a
// deterministic script that drives the browser through a full game lifecycle
// and records the entire session as a video.
//
// Output: docs/assets/demo-frontend.webm (raw recording, gitignored)
// The companion shell wrapper `scripts/demo-frontend.sh` converts it to a
// GIF and writes the final asset to docs/assets/demo-frontend.gif.
//
// Usage: scripts/demo-frontend.js
// Requires: @playwright/test installed, Chromium downloaded via
//           `npx playwright install chromium`, and a server running on :1337.

const { chromium } = require('@playwright/test');
const fs = require('fs');
const path = require('path');

const BASE_URL = process.env.DEMO_URL || 'http://localhost:1337';
const VIDEO_DIR = path.join(__dirname, '..', 'videos');
const FINAL_WEBM = path.join(__dirname, '..', 'docs', 'assets', 'demo-frontend.webm');

// Fixed word for the demo. Every run plays the same game so the GIF is
// byte-for-byte reproducible.
const DEMO_WORD = 'APPLE';
const DEMO_GAME_ID = 'demo-game-1';

const gameState = {
  current: '_____',
  guessesRemaining: 6,
  finished: false,
  won: false,
  lost: false,
  gameNumber: 0,
};

function newGameResponse() {
  gameState.gameNumber += 1;
  gameState.current = '_____';
  gameState.guessesRemaining = 6;
  gameState.finished = false;
  gameState.won = false;
  gameState.lost = false;
  return {
    id: `demo-game-${gameState.gameNumber}`,
    current: gameState.current,
    guesses_remaining: gameState.guessesRemaining,
  };
}

function guessResponse(letter) {
  if (gameState.finished) {
    return { error: 'game not found' };
  }

  const upper = letter.toUpperCase();
  const inWord = DEMO_WORD.includes(upper);

  if (inWord) {
    const revealed = gameState.current
      .split('')
      .map((c, i) => (DEMO_WORD[i] === upper ? upper : c))
      .join('');
    gameState.current = revealed;
    if (revealed === DEMO_WORD) {
      gameState.finished = true;
      gameState.won = true;
    }
  } else {
    gameState.guessesRemaining -= 1;
    if (gameState.guessesRemaining === 0) {
      gameState.finished = true;
      gameState.lost = true;
    }
  }

  const response = {
    id: `demo-game-${gameState.gameNumber}`,
    current: gameState.current,
    guesses_remaining: gameState.guessesRemaining,
  };
  if (gameState.won || gameState.lost) {
    response.word = DEMO_WORD;
  }
  return response;
}

async function main() {
  fs.mkdirSync(VIDEO_DIR, { recursive: true });
  fs.mkdirSync(path.dirname(FINAL_WEBM), { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1200, height: 800 },
    recordVideo: { dir: VIDEO_DIR, size: { width: 1200, height: 800 } },
  });
  const page = await context.newPage();

  // Intercept API calls with a deterministic state machine. The Go server
  // is still running on :1337, but we never let its responses reach the
  // page — our route handlers always win. This keeps the demo reproducible
  // without touching the production code.
  //
  // We use a single catch-all handler and dispatch by URL/method, because
  // Playwright's `page.route` matches in registration order and the more
  // specific patterns (POST /new, POST /guess) would otherwise be shadowed
  // by the abort catch-all.
  await page.route('**/api/v1/**', async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname;
    const method = route.request().method();

    if (method === 'POST' && path === '/api/v1/new') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(newGameResponse()),
      });
      return;
    }

    if (method === 'POST' && path === '/api/v1/guess') {
      const postData = route.request().postData() || '{}';
      const body = JSON.parse(postData);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(guessResponse(body.guess)),
      });
      return;
    }

    if (method === 'GET' && path.startsWith('/api/v1/game/')) {
      // useQuery re-fetches the game after newGame sets the cache. The
      // data is already in cache, but the queryFn runs anyway. Serve the
      // current snapshot so the network doesn't error.
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: `demo-game-${gameState.gameNumber}`,
          current: gameState.current,
          guesses_remaining: gameState.guessesRemaining,
          ...(gameState.finished ? { word: DEMO_WORD } : {}),
        }),
      });
      return;
    }

    // Any other /api/v1/* path is unexpected. Abort so the page shows a
    // real error rather than silently receiving a mock.
    await route.abort();
  });

  // Capture the video path up front. Playwright assigns it at context
  // creation, so we can read it before any game-flow step could throw.
  // The finally block below finalizes the context (flushing the .webm)
  // and copies the file even on failure, so a partial recording is
  // always preserved on disk for ffmpeg to pick up.
  let videoPath;
  try {
    videoPath = await page.video().path();
  } catch (_) {
    videoPath = null; // fall back: read after close in finally
  }

  try {
  console.log(`→ Opening ${BASE_URL}`);
  await page.goto(BASE_URL, { waitUntil: 'networkidle' });
  await page.waitForSelector('text=New Game', { timeout: 5000 });

  // -------- Game 1: win --------
  console.log('→ Click "New Game"');
  await page.click('button:has-text("New Game")');
  await page.waitForSelector('text=Guesses remaining: 6', { timeout: 5000 });

  // NOTE: one click per unique reveal letter. guessResponse("P") reveals
  // BOTH P positions in a single call (hangman-style, not Wordle), so we
  // do not click P twice — a second click would target a DOM-disabled
  // button (force:true can't override disabled) and is a silent no-op.
  // Scope the click to the keyboard key class. A bare
  // `button:has-text("L")` collides with the SiteTopNav theme-toggle
  // button (label "☀️ Light" contains capital "L"), so Playwright clicks
  // the toggle instead of the keyboard, the guess never fires, the L
  // position stays "_", and "You Won!" never renders. Same latent bug
  // for the loss loop: `has-text("D")` matches "🌙 Dark".
  for (const letter of ['A', 'P', 'L', 'E']) {
    console.log(`→ Click "${letter}"`);
    await page.click(`.game-page__key:has-text("${letter}")`);
    await page.waitForTimeout(700);
  }

  await page.waitForSelector('text=You Won!', { timeout: 5000 });
  await page.waitForTimeout(800);

  // -------- Game 2: loss --------
  console.log('→ Click "Play Again"');
  await page.click('button:has-text("Play Again")');
  await page.waitForSelector('text=Guesses remaining: 6', { timeout: 5000 });

  for (const letter of ['B', 'C', 'D', 'F', 'G', 'H']) {
    console.log(`→ Click "${letter}"`);
    await page.click(`.game-page__key:has-text("${letter}")`);
    await page.waitForTimeout(700);
  }

  await page.waitForSelector('text=Game Over', { timeout: 5000 });
  await page.waitForTimeout(1500);
  } finally {
    // Always finalize the .webm — even if the game flow threw. Without
    // context.close() Playwright never flushes the recording, leaving
    // the ffmpeg step with no input file.
    if (!videoPath) {
      try { videoPath = await page.video().path(); } catch (_) {}
    }
    await context.close();
    await browser.close();
  }

  fs.copyFileSync(videoPath, FINAL_WEBM);
  console.log(`✓ Wrote ${FINAL_WEBM}`);
  console.log(`  Next step: run scripts/demo-frontend.sh to convert to GIF`);
}

main().catch((err) => {
  console.error('demo-frontend.js failed:', err);
  process.exit(1);
});
