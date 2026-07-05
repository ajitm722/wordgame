import "@testing-library/jest-dom";
import { render, screen } from "test/test-utils";
import { App } from "./App";

describe("App", () => {
  test("renders idle game page with New Game button", () => {
    render(<App />);
    expect(screen.getByRole("button", { name: /new game/i })).toBeInTheDocument();
  });
});
