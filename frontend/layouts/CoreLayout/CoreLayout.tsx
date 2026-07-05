import { Outlet } from "react-router-dom";
import { SiteTopNav } from "layouts/SiteTopNav";

export function CoreLayout() {
  const baseClass = "core-layout";

  return (
    <div className={baseClass}>
      <SiteTopNav />
      <main className={`${baseClass}__content`}>
        <Outlet />
      </main>
    </div>
  );
}
