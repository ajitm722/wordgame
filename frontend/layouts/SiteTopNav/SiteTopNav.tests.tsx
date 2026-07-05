import "@testing-library/jest-dom";
import { render, screen } from "test/test-utils";
import { SiteTopNav } from "./SiteTopNav";

describe("SiteTopNav", () => {
  test("renders brand text", () => {
    render(<SiteTopNav />);
    expect(screen.getByText("Word Game")).toBeInTheDocument();
  });

  test("renders theme toggle button", () => {
    render(<SiteTopNav />);
    expect(screen.getByRole("button")).toBeInTheDocument();
  });
});
