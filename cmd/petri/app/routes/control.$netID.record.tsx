import { Outlet } from "@remix-run/react";


export default function ControlSystemPage() {
  return (
    <div>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
;