import endpoints from "utilities/endpoints";
import { sendRequest } from "utilities/sendRequest";
import type { INewGameResponse, IGuessResponse } from "interfaces/game";

export default {
  newGame: () => sendRequest<INewGameResponse>("POST", endpoints.NEW_GAME),

  getGame: (id: string) =>
    sendRequest<IGuessResponse>("GET", endpoints.GAME_BY_ID(id)),

  guess: (id: string, guess: string) =>
    sendRequest<IGuessResponse>("POST", endpoints.GUESS, { id, guess }),
};
