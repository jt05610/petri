import { requireUserId } from "~/session.server";
import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { listInstances } from "~/models/device.instance.server";
import { NavLink, Outlet, useLoaderData } from "@remix-run/react";


export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.deviceID, "deviceID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const instances = await listInstances({ deviceID: params.deviceID });
  return json({ instances: instances });
};

export default function Instances() {
  const data = useLoaderData<typeof loader>();
  return (
    <div>
      <div className={"flex flex-col items-center justify-start space-y-1 p-2"}>
        <h3 className="text-2xl font-bold">Instances</h3>
        {data.instances.map((instance) => (
          <div className={`flex flex-row justify-start w-full`} key={instance.id}>
            <NavLink to={instance.id}
                     key={instance.id}>
              {instance.name}
            </NavLink>
          </div>
        ))}
      </div>
      <Outlet />
    </div>
  );
};