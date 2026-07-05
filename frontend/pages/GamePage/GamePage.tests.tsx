import "@testing-library/jest-dom";
import { render, screen, waitFor } from "test/test-utils";
import userEvent from "@testing-library/user-event";
import { GamePage } from "./GamePage";
import { http, HttpResponse } from "msw";
import mockServer from "test/mock-server";

function renderPage() {
  return render(<GamePage />);
}

describe("GamePage", () => {
  test("renders idle state with New Game button", () => {
    renderPage();
    expect(screen.getByText("Word Game")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /new game/i })).toBeInTheDocument();
  });

  test("starts a game when New Game is clicked", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));

    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
    });

    expect(screen.getAllByText("_")).toHaveLength(5);
  });

  test("renders keyboard in playing state", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));

    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
    });

    for (const letter of "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
      expect(screen.getByRole("button", { name: letter })).toBeInTheDocument();
    }
  });

  test("disables guessed letter and updates remaining guesses", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));

    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Z" }));

    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 5")).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: "Z" })).toBeDisabled();
  });

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

    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));
    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "A" }));

    await waitFor(() => {
      expect(screen.getByText("You Won!")).toBeInTheDocument();
    });

    expect(screen.getByText("You Won!")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /play again/i })).toBeInTheDocument();
  });

  test("renders lost state with secret word", async () => {
    mockServer.use(
      http.post("/api/v1/guess", () => {
        return HttpResponse.json({
          id: "test-id",
          current: "_____",
          guesses_remaining: 0,
          word: "APPLE",
        });
      })
    );

    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));
    await waitFor(() => {
      expect(screen.getByText("Guesses remaining: 6")).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Z" }));

    await waitFor(() => {
      expect(screen.getByText("Game Over")).toBeInTheDocument();
    });

    expect(screen.getByText("APPLE")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
  });

  test("renders error state with retry button", async () => {
    mockServer.use(
      http.post("/api/v1/new", () => {
        return HttpResponse.json(
          { message: "Server error" },
          { status: 500 }
        );
      })
    );

    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /new game/i }));

    await waitFor(() => {
      expect(screen.getByText(/server error/i)).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
  });
});
