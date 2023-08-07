import { Form, Outlet } from "@remix-run/react";

export default function ControlRoute() {
  return (
    <div>
      <header className="flex items-center justify-between bg-slate-800 p-4 text-white">
        <div className={"flex flex-row"}>
        </div>
        <div>
          <Form action="/logout" method="post">
            <button
              type="submit"
              className="rounded bg-slate-600 px-4 py-2 text-blue-100 hover:bg-blue-500 active:bg-blue-600"
            >
              Logout
            </button>
          </Form>
        </div>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}