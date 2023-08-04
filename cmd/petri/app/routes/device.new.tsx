import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { createDevice, DeviceInputSchema } from "~/models/device.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/components/form";
import { useState } from "react";
import { getNetListItems } from "~/models/net.server";

export const action = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  let formData = await request.formData();
  const submission = parse(formData, {
    schema: DeviceInputSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  const device = await createDevice(submission.value);
  return redirect(`/device/${device.id}`);
};

export const loader = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const nets = await getNetListItems({ authorID });
  return json({ nets });
};

export default function Transition() {
  const lastSubmission = useActionData<typeof action>();
  const { nets } = useLoaderData<typeof loader>();
  const [changed, setChanged] = useState(false);
  const [form, { name, description, netID }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      setChanged(false);
      return parse(formData, { schema: DeviceInputSchema });
    }
  });

  return (
    <div>
      <div className={"rounded-lg border-2 p-2 "}>
        <h2 className={"text-lg font-semibold"}>Update</h2>
        <Form method={"post"} {...form.props} onChange={() => setChanged(true)}>
          <FormContent activeButton={changed} fields={[
            {
              type: "text",
              name: name.name,
              content: name.form,
              error: name.error
            },
            {
              name: description.name,
              type: "textarea",
              content: description.form,
              error: description.error
            },
            {
              name: netID.name,
              type: "select",
              content: lastSubmission?.payload?.transitionID,
              error: netID.error,
              options: nets.map((net) => ({ display: net.name, value: net.id }))
            }
          ]} />
        </Form>
      </div>
      <div className={"rounded-lg border-2 p-2 "}>
        <h2 className={"text-lg font-semibold"}>Events</h2>
        <Form method={"post"} {...form.props} onChange={() => setChanged(true)}>
          <button
            className={"bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"}
            onClick={() => setChanged(true)}
          >Add Event
          </button>
        </Form>
      </div>
    </div>
  );

}