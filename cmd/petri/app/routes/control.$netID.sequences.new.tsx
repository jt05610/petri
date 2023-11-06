import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { addRun } from "~/models/net.run.server";
import { parse } from "@conform-to/zod";
import { RunInputSchema } from "~/models/net.run";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import invariant from "tiny-invariant";
import { Field, FieldSetTable, FieldTextArea } from "~/lib/components/FormFieldSet";
import React from "react";
import { Form, useActionData } from "@remix-run/react";
import { useFieldList, useForm } from "@conform-to/react";

export const action = async ({ params, request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  const netID = params.netID;
  invariant(netID, "netID is required");
  const formData = await request.formData();

  const submission = parse(formData, { schema: RunInputSchema });
  if (submission.intent !== "submit" || !submission.value) {
    console.log("invalid submission", submission);
    return json(submission);
  }
  const sequence = await addRun(netID, submission.value);
  return redirect(`/control/${params.netID}/sequences/${sequence.id}/record`);
};

export default function NewSequence() {
  const lastSubmission = useActionData();
  const [form, { name, parameters, description }] = useForm({
    lastSubmission,
    shouldValidate: "onBlur",
    onValidate({ formData }) {
      const sub = parse(formData, { schema: RunInputSchema });
      console.log("sub", sub);
      return sub;
    }
  });
  const parameterList = useFieldList(form.ref, parameters);
  return (
    <Form
      method="post"
      {...form.props}
      className={"flex flex-col w-96 space-2"}
    >
      <Field {...name} />
      <FieldTextArea {...description} />
      <h2
        className={"text-lg font-semibold text-center text-slate-900 dark:text-slate-100"}
      >
        Parameters
      </h2>
      <FieldSetTable config={parameters} fieldList={parameterList} fields={["name", "expression"]} />
      <button
        type="submit"
        className={`px-2 py-1 text-white rounded-full bg-teal-900 hover:text-sky-500 dark:hover:text-sky-400`}
      >
        Record
      </button>
    </Form>
  );
}