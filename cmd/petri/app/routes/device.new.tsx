import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import type { DeviceInput } from "~/models/device.server";
import { createDevice, DeviceInputSchema } from "~/models/device.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useFieldList, useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { getNetListItems } from "~/models/net.server";

export const action = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  let formData = await request.formData();
  formData.set("authorID", authorID);
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
  return json({ nets: nets });
};

export default function Transition() {
  const lastSubmission = useActionData<typeof action>();
  const { nets } = useLoaderData<typeof loader>();
  const [form, { name, description, netIDs }] = useForm<DeviceInput>({
    lastSubmission
  });
  const netIDList = useFieldList(form.ref, netIDs);
  return (
    <div>
      <h2 className={"text-lg font-semibold"}>Update</h2>
      <Form method={"post"} {...form.props}>
        <FormContent fields={[
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
            name: netIDs.name,
            type: "multiselect",
            content: netIDList,
            error: netIDs.error,
            options: nets.map((net) => ({ display: net.name, value: net.id }))
          }
        ]} />
      </Form>
    </div>
  );

}