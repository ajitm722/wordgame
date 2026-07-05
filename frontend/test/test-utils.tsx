import React, { type ReactElement } from "react";
import { render, type RenderOptions } from "@testing-library/react";
import { AppProvider } from "context/app";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

/*
 * Custom render — wraps the component under test with the same
 * providers that the real App uses (AppProvider + QueryProvider),
 * so tests don't need to duplicate provider setup.
 * Creates a fresh QueryClient per render for test isolation.
 */
function customRender(
  ui: ReactElement,
  options?: Omit<RenderOptions, "wrapper">
) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, refetchOnMount: false },
      mutations: { retry: false },
    },
  });

  function AllProviders({ children }: { children: React.ReactNode }) {
    return (
      <AppProvider>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </AppProvider>
    );
  }

  return render(ui, { wrapper: AllProviders, ...options });
}

export { customRender as render };

export { screen, waitFor } from "@testing-library/react";
