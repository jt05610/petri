import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { addTransition } from "~/models/net.server";
import invariant from "tiny-invariant";

export const action = async ({ params, request }: ActionArgs) => {
  const formData = await request.formData();
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  const name = formData.get("name");
  const description = formData.get("description");
  const condition = formData.get("condition");
  if (typeof name !== "string" || name.length === 0) {
    return json(
      { errors: { body: null, title: "Title is required" } },
      { status: 400 }
    );
  }
  if (typeof description !== "string") {
    return json(
      { errors: { body: null, title: "Description is required" } },
      { status: 400 }
    );
  }
  if (typeof condition !== "string") {
    return json(
      { errors: { body: null, title: "Condition is required" } },
      { status: 400 }
    );
  }
  const res = await addTransition({ name: name, description: description, condition: condition, netID: netID });
  return json(res);
};