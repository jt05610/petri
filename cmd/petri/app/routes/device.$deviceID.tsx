import { NavLink, Outlet, useLoaderData } from "@remix-run/react";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getDevice } from "~/models/device.server";
import { getUserById } from "~/models/user.server";

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  invariant(params.deviceID, "deviceID not found");
  invariant(user, "User not found");

  const device = await getDevice({ id: params.deviceID });
  if (!device) {
    throw new Response("Not Found", { status: 404 });
  }

  return json({ device });
};


export default function DeviceDetail() {
  const { device } = useLoaderData<typeof loader>();
  return (
    <div
      className={"flex flex-col w-full h-full space-y-2 p-2 justify-between"}
    >
      <div className={"flex flex-row w-full h-full"}>
        <Outlet />
      </div>
      <div className={"flex flex-col w-full h-full"}>
      </div>
      <div className={"flex flex-row w-full sticky bottom-0"}>
        <div className={"flex flex-col"}>
          <h1 className={"font-bold text-2xl"}>
            {device.name}
          </h1>
          <p className={`${device.description ? "" : "hidden"}`}>
            {device.description}
          </p>
        </div>
        <div className={"flex flex-row space-x-2"}>
          <div className={"flex flex-col rounded-lg bg-slate-100 space-y-1 p-1"}>
            <h3 className="text-2xl font-bold">Actions</h3>
            <hr className="rounded-full border-2 border-slate-200" />
            <NavLink to={`/device/${device.id}/edit`}>Edit</NavLink>
          </div>
          <div className={"flex flex-col rounded-lg bg-slate-100 space-y-1 p-1"}>
            <h3 className="text-2xl font-bold">Instances</h3>
            <hr className="rounded-full border-2 border-slate-200" />
            {device.instances.map((instance, i) => (
              <div key={instance.id}>
                <NavLink to={`instance/${instance.id}`}>{i + 1}. {instance.name}</NavLink>
              </div>
            ))}
            <NavLink
              className={"flex rounded bg-slate-600 text-white px-2 py-1"}
              to={`instance/new`}>New</NavLink>
          </div>
        </div>
      </div>
    </div>
  );

}