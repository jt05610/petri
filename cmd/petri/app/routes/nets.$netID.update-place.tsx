import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import {updatePlace } from "~/models/net.server";

export const action = async ({  request }: ActionArgs) => {
  const formData = await request.formData();
  const id = formData.get("id");
  const name = formData.get("name");
  const description = formData.get("description");
  const bound = formData.get("bound");
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

  const res = await updatePlace({ id: id, name: name, description: description, bound: Number(bound) });
  return json(res);
};