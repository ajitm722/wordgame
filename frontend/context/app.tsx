import React, { createContext, useReducer, useContext, type ReactNode } from "react";

interface IAppState {
  theme: "light" | "dark";
  config: {
    apiBaseUrl: string;
  };
}

const initialState: IAppState = {
  theme: "dark",
  config: {
    apiBaseUrl: "http://localhost:1337",
  },
};

type AppAction = { type: "SET_THEME"; payload: "light" | "dark" };

function appReducer(state: IAppState, action: AppAction): IAppState {
  switch (action.type) {
    case "SET_THEME":
      return { ...state, theme: action.payload };
    default:
      return state;
  }
}

interface IAppContext {
  state: IAppState;
  dispatch: React.Dispatch<AppAction>;
}

const AppCtx = createContext<IAppContext | null>(null);

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState);
  return <AppCtx.Provider value={{ state, dispatch }}>{children}</AppCtx.Provider>;
}

export function useAppContext(): IAppContext {
  const ctx = useContext(AppCtx);
  if (!ctx) throw new Error("useAppContext must be used within AppProvider");
  return ctx;
}
