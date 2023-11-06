import type { ActionArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { Form, useActionData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { updateNet } from "~/models/net.server";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { UpdateNetSchema } from "~/models/net";
import { useForm } from "@conform-to/react";
import { Field, FieldList, FieldTextArea, SelectField } from "~/lib/components/FormFieldSet";
import React from "react";
import { useNet } from "~/lib/context/NetContext";

export const action = async ({ params, request }: ActionArgs) => {
  const authorID = await requireUserId(request);
  const netID = params.netID;
  invariant(netID, "netID is required");
  const user = await getUserById(authorID);
  invariant(user, "user is required");
  const formData = await request.formData();
  const submission = parse(formData, { schema: UpdateNetSchema });
  if (submission.intent !== "submit" || !submission.value) {
    return json(submission);
  }
  const net = await updateNet(netID, submission.value);
  return redirect(`/nets/${net.id}`);
};

export default function EditNetPage() {
  const lastSubmission = useActionData<typeof action>();
  const net = useNet();
  const [form, { name, description, visibility, sharedWith }] = useForm({
    lastSubmission,
    shouldValidate: "onBlur",
    onValidate({ formData }) {
      return parse(formData, { schema: UpdateNetSchema });
    }
  });
  return (
    <Form
      method="post"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 8,
        width: "100%"
      }}
      {...form.props}
    >
      <Field {...name} defaultValue={net.name} />
      <FieldTextArea {...description} defaultValue={net.description} />
      <SelectField {...visibility} options={["PRIVATE", "PUBLIC"]} defaultValue={net.visibility} />
      <FieldList {...sharedWith} defaultValue={net.sharedWith.map((user) => user.user.email)} />
      <button
        type="submit"
        className={"block rounded-full w-full px-3 py-2 border border-gray-300 shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:bg-slate-800 dark:border-slate-400 dark:text-gray-300"}
      >Submit
      </button>
    </Form>
  );
}
