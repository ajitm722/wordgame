const API_BASE = "/api/v1";

export default {
  NEW_GAME: `${API_BASE}/new`,
  GUESS: `${API_BASE}/guess`,
  GAME_BY_ID: (id: string) => `${API_BASE}/game/${id}`,
};
