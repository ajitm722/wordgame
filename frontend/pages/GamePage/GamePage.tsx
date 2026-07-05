import { useGame } from "hooks/useGame";

/*
 * GamePage — the main game screen.
 * Renders one of four states:
 *   1. Idle:    no game in progress, shows "New Game" button
 *   2. Playing: board tiles + keyboard + guess counter
 *   3. Won:     congratulations message + revealed word
 *   4. Lost:    game over message + the secret word
 *   5. Error:   error message + retry button
 */
export function GamePage() {
  const baseClass = "game-page";

  const {
    current,
    word,
    guessesRemaining,
    newGame,
    makeGuess,
    isLoading,
    guessedLetters,
    isPendingGuess,
    isWon,
    isLost,
    isError,
    error,
  } = useGame();

  if (isError) {
    return (
      <div className={baseClass}>
        <p className={`${baseClass}__error`}>Error: {error?.message}</p>
        <button className={`${baseClass}__btn`} onClick={newGame}>
          Try Again
        </button>
      </div>
    );
  }

  // Idle state — no game started yet
  if (current === null) {
    return (
      <div className={baseClass}>
        <h1 className={`${baseClass}__title`}>Word Game</h1>
        <button
          className={`${baseClass}__btn ${baseClass}__btn--primary`}
          onClick={newGame}
          disabled={isLoading}
        >
          {isLoading ? "Starting..." : "New Game"}
        </button>
      </div>
    );
  }

  // Terminal states
  if (isWon) {
    return (
      <div className={baseClass}>
        <h1 className={`${baseClass}__title`}>You Won!</h1>
        <div className={`${baseClass}__board`}>
          {current.split("").map((letter, i) => (
            <span key={i} className={`${baseClass}__tile ${baseClass}__tile--revealed`}>
              {letter}
            </span>
          ))}
        </div>
        <button
          className={`${baseClass}__btn ${baseClass}__btn--primary`}
          onClick={newGame}
        >
          Play Again
        </button>
      </div>
    );
  }

  if (isLost) {
    return (
      <div className={baseClass}>
        <h1 className={`${baseClass}__title`}>Game Over</h1>
        <p className={`${baseClass}__message`}>
          The word was: <strong>{word ?? current}</strong>
        </p>
        <button
          className={`${baseClass}__btn ${baseClass}__btn--primary`}
          onClick={newGame}
        >
          Try Again
        </button>
      </div>
    );
  }

  // Playing state
  return (
    <div className={baseClass}>
      <h1 className={`${baseClass}__title`}>Word Game</h1>

      <div className={`${baseClass}__board`}>
        {current.split("").map((letter, i) => (
          <span
            key={i}
            className={`${baseClass}__tile ${
              letter !== "_" ? `${baseClass}__tile--revealed` : ""
            }`}
          >
            {letter}
          </span>
        ))}
      </div>

      <div className={`${baseClass}__keyboard`}>
        {"ABCDEFGHIJKLMNOPQRSTUVWXYZ".split("").map((letter) => (
          <button
            key={letter}
            className={`${baseClass}__key`}
            onClick={() => makeGuess(letter)}
            disabled={isPendingGuess || guessedLetters.includes(letter)}
          >
            {letter}
          </button>
        ))}
      </div>

      <p className={`${baseClass}__guesses`}>
        Guesses remaining: {guessesRemaining}
      </p>

      <button className={`${baseClass}__btn`} onClick={newGame}>
        New Game
      </button>
    </div>
  );
}
