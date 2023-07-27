import type { ActionArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";

import { addPlace } from "~/models/net.server";

export const action = async ({ request }: ActionArgs) => {
  const formData = await request.formData();
  const name = formData.get("name");
  const bound = formData.get("bound");
  const netID = formData.get("netID");

  if (typeof name !== "string" || name.length === 0) {
    return json(
      { errors: { body: null, title: "Title is required" } },
      { status: 400 }
    );
  }
  if (typeof bound !== "string" || bound.length === 0) {
    return json(
      { errors: { body: null, title: "Bound is required" } },
      { status: 400 }
    );
  }

  if (typeof netID !== "string" || netID.length === 0) {
    return json(
      { errors: { body: null, title: "NetID is required" } },
      { status: 400 }
    );
  }
  // convert bound to number
  await addPlace({ name: name, bound: Number(bound), netID: netID });
  return redirect(`/nets/${netID}`);
};

export const loader = async ({ request }: ActionArgs) => {
  const formData = await request.formData();
  const netID = formData.get("netID");
  return json({ netID: netID });
}
