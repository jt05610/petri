import type { ActionArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { Form, useActionData } from "@remix-run/react";
import { useEffect, useRef } from "react";

import { createNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";

export const action = async ({ request }: ActionArgs) => {
  const authorID = await requireUserId(request);

  const formData = await request.formData();
  const name = formData.get("title");
  const description = formData.get("desc");

  if (typeof name !== "string" || name.length === 0) {
    return json(
      { errors: { body: null, title: "Title is required" } },
      { status: 400 }
    );
  }

  if (typeof description !== "string" || description.length === 0) {
    return json(
      { errors: { body: "Description is required", title: null } },
      { status: 400 }
    );
  }

  const net = await createNet({ name, description, authorID });

  return redirect(`/nets/${net.id}`);
};

export default function NewNetPage() {
  const actionData = useActionData<typeof action>();
  const titleRef = useRef<HTMLInputElement>(null);
  const bodyRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (actionData?.errors?.title) {
      titleRef.current?.focus();
    } else if (actionData?.errors?.body) {
      bodyRef.current?.focus();
    }
  }, [actionData]);

  return (
    <Form
      method="post"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 8,
        width: "100%"
      }}
    >
      <div>
        <label className="flex w-full flex-col gap-1">
          <span>Name: </span>
          <input
            ref={titleRef}
            name="title"
            className="flex-1 rounded-full hover:border-2 focus:border-teal-400 hover:border-teal-500 px-3 text-lg leading-loose dark:bg-slate-800 dark:text-gray-200"
            aria-invalid={actionData?.errors?.title ? true : undefined}
            aria-errormessage={
              actionData?.errors?.title ? "title-error" : undefined
            }
          />
        </label>
        {actionData?.errors?.title ? (
          <div className="pt-1 text-red-700" id="title-error">
            {actionData.errors.title}
          </div>
        ) : null}
      </div>

      <div>
        <label className="flex w-full flex-col gap-1">
          <span>Description: </span>
          <textarea
            ref={bodyRef}
            name="desc"
            rows={8}
            className="w-full flex-1 rounded-xl hover:border-2 focus:border-teal-400 hover:border-teal-500 px-3 text-lg dark:bg-slate-800 dark:text-gray-200 leading-6"
            aria-invalid={actionData?.errors?.body ? true : undefined}
            aria-errormessage={
              actionData?.errors?.body ? "body-error" : undefined
            }
          />
        </label>
        {actionData?.errors?.body ? (
          <div className="pt-1 text-red-700" id="body-error">
            {actionData.errors.body}
          </div>
        ) : null}
      </div>

      <div className="text-right">
        <button
          type="submit"
          className="rounded-full bg-teal-500 px-4 py-2 text-white hover:bg-teal-600 focus:bg-teal-400"
        >
          Save
        </button>
      </div>
    </Form>
  );
}
