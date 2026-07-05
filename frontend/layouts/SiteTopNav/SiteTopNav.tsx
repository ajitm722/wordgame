import { useAppContext } from "context/app";

export function SiteTopNav() {
  const baseClass = "site-top-nav";
  const { state, dispatch } = useAppContext();

  function toggleTheme() {
    dispatch({
      type: "SET_THEME",
      payload: state.theme === "light" ? "dark" : "light",
    });
  }

  return (
    <nav className={baseClass}>
      <span className={`${baseClass}__title`}>Word Game</span>
      <button className={`${baseClass}__theme-btn`} onClick={toggleTheme}>
        {state.theme === "light" ? "🌙 Dark" : "☀️ Light"}
      </button>
    </nav>
  );
}
