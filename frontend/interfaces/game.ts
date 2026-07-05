export interface INewGameResponse {
  id: string;
  current: string;
  guesses_remaining: number;
}

export interface IGuessResponse {
  id: string;
  current: string;
  guesses_remaining: number;
  word?: string;
}

export interface IGameError {
  message: string;
  code?: string;
}
