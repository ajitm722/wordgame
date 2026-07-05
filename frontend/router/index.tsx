import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { CoreLayout } from "layouts/CoreLayout";
import { GamePage } from "pages/GamePage";
import PATHS from "./paths";

/*
 * React Router v6 route tree — uses createBrowserRouter for
 * data-loader-style routing. All game routes live under CoreLayout.
 */
const router = createBrowserRouter([
  {
    path: PATHS.HOME,
    element: <CoreLayout />,
    children: [{ index: true, element: <GamePage /> }],
  },
]);

export function AppRouter() {
  return <RouterProvider router={router} />;
}
