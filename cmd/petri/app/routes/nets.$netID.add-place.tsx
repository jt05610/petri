import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { addPlace } from "~/models/net.server";
import invariant from "tiny-invariant";

export const action = async ({ params, request }: ActionArgs) => {
  const formData = await request.formData();
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  const name = formData.get("name");
  const description = formData.get("description");
  const bound = formData.get("bound");
  if (typeof name !== "string" || name.length === 0) {
    return json(
      { errors: { body: null, title: "Title is required" } },
      { status: 400 }
    );
  }
  if (typeof description !== "string" || description.length === 0) {
    return json(
      { errors: { body: null, title: "Description is required" } },
      { status: 400 }
    );
  }
  if (typeof bound !== "string" || bound.length === 0) {
    return json(
      { errors: { body: null, title: "Bound is required" } },
      { status: 400 }
    );
  }

  const res = await addPlace({ name: name, description: description, bound: Number(bound), netID: netID });
  return json(res);
};