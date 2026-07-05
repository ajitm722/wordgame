import { useEffect } from "react";
import { AppProvider, useAppContext } from "context/app";
import { QueryProvider } from "context/query";
import { AppRouter } from "router";

function ThemeSync() {
  const { state } = useAppContext();
  useEffect(() => {
    document.documentElement.dataset.theme = state.theme;
  }, [state.theme]);
  return null;
}

export function App() {
  return (
    <AppProvider>
      <ThemeSync />
      <QueryProvider>
        <AppRouter />
      </QueryProvider>
    </AppProvider>
  );
}
