import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, useCallback } from "react";
import gameAPI from "services/entities/game";
import type { IGuessResponse } from "interfaces/game";

export function useGame() {
  const queryClient = useQueryClient();
  const [gameId, setGameId] = useState<string | null>(null);
  const [guessedLetters, setGuessedLetters] = useState<string[]>([]);

  const { data: gameData } = useQuery<IGuessResponse>({
    queryKey: ["game", gameId],
    queryFn: () => gameAPI.getGame(gameId!),
    enabled: !!gameId,
  });

  const newGameMutation = useMutation({
    mutationFn: gameAPI.newGame,
    onSuccess: (data) => {
      queryClient.setQueryData(["game", data.id], data);
      setGameId(data.id);
      setGuessedLetters([]);
    },
    onError: () => {
      setGameId(null);
      setGuessedLetters([]);
    },
  });

  const guessMutation = useMutation({
    mutationFn: ({ id, guess }: { id: string; guess: string }) =>
      gameAPI.guess(id, guess),
    onSuccess: (data, variables) => {
      queryClient.setQueryData(["game", variables.id], data);
    },
    onError: () => {
      setGameId(null);
      setGuessedLetters([]);
    },
  });

  const current = gameData?.current ?? null;
  const guessesRemaining = gameData?.guesses_remaining ?? null;
  const word = gameData?.word ?? null;

  const newGame = useCallback(() => {
    newGameMutation.mutate();
  }, [newGameMutation]);

  const makeGuess = useCallback(
    async (letter: string) => {
      if (!gameId) return;
      setGuessedLetters((prev) => [...prev, letter]);
      await guessMutation.mutateAsync({ id: gameId, guess: letter });
    },
    [gameId, guessMutation]
  );

  const isWon = current !== null && !current.includes("_");
  const isLost = current !== null && guessesRemaining !== null && guessesRemaining <= 0;

  return {
    gameId,
    current,
    word,
    guessesRemaining,
    guessedLetters,
    newGame,
    makeGuess,
    isWon,
    isLost,
    isLoading: newGameMutation.isPending,
    isPendingGuess: guessMutation.isPending,
    isError: newGameMutation.isError || guessMutation.isError,
    error: newGameMutation.error || guessMutation.error,
  };
}
