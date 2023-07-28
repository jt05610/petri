import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { updateTransition } from "~/models/net.server";

export const action = async ({ params, request }: ActionArgs) => {
  const formData = await request.formData();
  const name = formData.get("name");
  const description = formData.get("description");
  const condition = formData.get("condition");
  const id = formData.get("id");
  if (typeof id !== "string" || id.length === 0) {
    return json(
      { errors: { body: null, title: "ID is required" } },
      { status: 400 }
    );
  }
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
  const res = await updateTransition({ id: id, name: name, description: description, condition: condition });
  return json(res);
};