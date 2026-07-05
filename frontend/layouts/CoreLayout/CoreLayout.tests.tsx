import "@testing-library/jest-dom";
import { render, screen } from "test/test-utils";
import { CoreLayout } from "./CoreLayout";

describe("CoreLayout", () => {
  test("renders SiteTopNav", () => {
    render(<CoreLayout />);
    expect(screen.getByText("Word Game")).toBeInTheDocument();
  });
});
