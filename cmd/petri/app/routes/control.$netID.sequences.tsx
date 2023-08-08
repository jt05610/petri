import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { Link, NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { listRuns } from "~/models/net.run.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import Header from "~/lib/components/header";
import invariant from "tiny-invariant";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  invariant(params.netID, "netID not found");
  const sequences = await listRuns({ netID: params.netID });

  return json({ sequences });
};

export default function RunsPage() {
  const { sequences } = useLoaderData<typeof loader>();
  return (
    <div className="flex h-full min-h-screen flex-col">
      <Header />
      <main className="flex h-full bg-white">
        <div className="h-full w-80 border-r bg-gray-50">
          <Link to="admin" className="block p-4 text-xl text-blue-500">
            Admin
          </Link>

          <hr />
          {sequences.length === 0 ? (
            <p className="p-4">No sequences yet</p>
          ) : (
            <ol>
              {sequences.map((sequence) => (
                <li key={sequence.id}>
                  <NavLink
                    className={({ isActive }) =>
                      `block border-b p-4 text-xl ${isActive ? "bg-white" : ""}`
                    }
                    to={sequence.id}
                  >
                    {sequence.name}
                  </NavLink>
                </li>
              ))}
            </ol>
          )}
        </div>

        <div className="flex-1 p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
