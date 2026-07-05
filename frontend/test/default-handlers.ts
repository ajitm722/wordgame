import { http, HttpResponse } from "msw";

export const handlers = [
  http.post("/api/v1/new", () => {
    return HttpResponse.json({
      id: "test-id",
      current: "_____",
      guesses_remaining: 6,
    });
  }),

  http.get("/api/v1/game/:id", ({ params }) => {
    return HttpResponse.json({
      id: params.id,
      current: "_____",
      guesses_remaining: 6,
    });
  }),

  http.post("/api/v1/guess", () => {
    return HttpResponse.json({
      id: "test-id",
      current: "_____",
      guesses_remaining: 5,
    });
  }),
];

export default handlers;
