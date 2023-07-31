import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { Link, NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { getNetListItems } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import { UserContext } from "~/context/user";
import { getUserById } from "~/models/user.server";
import Header from "~/lib/components/header";
import { SpaceContext } from "~/context/space";

export const loader = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const netLIstItems = await getNetListItems({ authorID });
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  return json({ netListItems: netLIstItems, user });
};

export default function NetsPage() {
  const data = useLoaderData<typeof loader>();
  return (
    <SpaceContext.Provider value={"Nets"}>
      <UserContext.Provider value={data.user}>
        <div className="flex h-full min-h-screen flex-col">
          <Header />
          <main className="flex h-full bg-white">
            <div className="h-full w-80 border-r bg-gray-50">
              <Link to="admin" className="block p-4 text-xl text-blue-500">
                Admin
              </Link>

              <hr />

              {data.netListItems.length === 0 ? (
                <p className="p-4">No notes yet</p>
              ) : (
                <ol>
                  {data.netListItems.map((net) => (
                    <li key={net.id}>
                      <NavLink
                        className={({ isActive }) =>
                          `block border-b p-4 text-xl ${isActive ? "bg-white" : ""}`
                        }
                        to={net.id}
                      >
                        𐄳 {net.name}
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
      </UserContext.Provider>
    </SpaceContext.Provider>
  );
}
