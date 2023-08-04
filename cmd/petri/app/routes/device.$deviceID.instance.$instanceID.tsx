import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { deleteInstance, getInstance } from "~/models/instance.server";
import { Form, NavLink, Outlet, useLoaderData } from "@remix-run/react";
import { useState } from "react";
import { sayHello } from "~/models/control.server";

export const action = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.deviceID, "deviceID not found");
  invariant(params.instanceID, "instanceID not found");
  const user = await getUserById(authorID);
  invariant(user, "User not found");
  invariant(params.instanceID, "instanceID not found");
  await deleteInstance({ id: params.instanceID, deviceID: params.deviceID });
  return redirect(`/device/${params.deviceID}/instance`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.deviceID, "deviceID not found");
  invariant(params.instanceID, "instanceID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  await sayHello({ deviceID: params.instanceID });
  const instance = await getInstance({ id: params.instanceID });
  return json({ instance });
};

export default function Instance() {
  const { instance } = useLoaderData<typeof loader>();
  const [alive, isAlive] = useState(false);
  const creation = new Date(instance.createdAt).toLocaleString();
  const updated = new Date(instance.updatedAt).toLocaleString();

  const healthCheck = async () => {
    const res = await fetch(`${instance.id}/hello`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ deviceID: instance.id })
    });
    const {success} = await res.json();
    isAlive(success);
  };

  return (
    <div className={"flex flex-col justify-center space-y-2 p-2"}>
      <h1 className={"text-2xl font-bold"}>{instance.addr}</h1>
      <h2 className={"text-xl"}>Status: {alive ? "Connected" : "Disconnected"}</h2>
      <h2 className={"text-xl"}>{instance.name}</h2>
      <h3 className={"text-md"}>language: {instance.language.toLowerCase()}</h3>
      <p className={"text-xs justify-start"}>ID: {instance.id}</p>
      <p className={"text-sm justify-start"}>Created: {creation}</p>
      <p className={"text-sm"}>Updated: {updated}</p>
      <br />
      <div className={"flex flex-row flex-shrink justify-between"}>
        <NavLink to={"edit"}
                 className={"rounded-full contents-center bg-slate-600 text-white justify-center flex px-2 py-1"}>
          Edit
        </NavLink>
        <button onClick={healthCheck}
                className={"rounded-full contents-center bg-black text-white justify-center flex px-2 py-1"}>
          HealthCheck
        </button>
        <Form method={"post"}>
          <button type={"submit"}
                  className={"rounded-full contents-center bg-red-800 text-white justify-center flex px-2 py-1"}>
            Delete
          </button>
        </Form>
      </div>
      <br />
      {instance.events && instance.events.map((event, i) => (
        <div className={"flex flex-col justify-start space-y-1 p-2"} key={i}>
          <h3 className="text-2xl font-bold">{event.name}</h3>
          <p className={"text-xs justify-start"}>{event.description}</p>
          <h3 className="text-lg font-bold">Fields</h3>
          <table className={"table-auto"}>
            <thead>
            <tr>
              <th className={"px-4 py-2"}>Type</th>
              <th className={"px-4 py-2"}>Name</th>
            </tr>
            </thead>
            <tbody>
            {event.fields.map((field, i) => (
              <tr key={i}>
                <td className={"border px-4 py-2"}>{field.type}</td>
                <td className={"border px-4 py-2"}>{field.name}</td>
              </tr>
            ))}
            </tbody>
          </table>
        </div>
      ))
      }
      <div>
        <Outlet />
      </div>
    </div>
  );
}