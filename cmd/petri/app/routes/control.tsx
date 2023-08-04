import { Outlet } from "@remix-run/react";
export default function ControlRoute() {

  return (
    <div>
      <h1>Control</h1>
      <main>
          <Outlet />
      </main>
    </div>
  );
}