import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { getDevice, updateDevice, UpdateDeviceInputSchema } from "~/models/device.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { useState } from "react";
import { getNetListItems } from "~/models/net.server";
import invariant from "tiny-invariant";

export const action = async ({ request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  let formData = await request.formData();
  formData.append("id", authorID);
  const submission = parse(formData, {
    schema: UpdateDeviceInputSchema
  });
  console.log(submission.value)
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }
  const device = await updateDevice(submission.value);
  return redirect(`/device/${device.id}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  invariant(params.deviceID, "deviceID not found");

  const device = await getDevice({ id: params.deviceID });
  const nets = await getNetListItems({ authorID });
  return json({ device, nets });
};

export default function Transition() {
  const lastSubmission = useActionData<typeof action>();
  const { device, nets } = useLoaderData<typeof loader>();
  const [changed, setChanged] = useState(false);
  const [form, { name, description, netIDs }] = useForm({
    lastSubmission,
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
              content: device.name,
              error: name.error
            },
            {
              name: description.name,
              type: "textarea",
              content: device.description,
              error: description.error
            },
            {
              name: netIDs.name,
              type: "multiselect",
              content: device.nets.map((net) => net.net.id),
              error: netIDs.error,
              options: nets.map((net) => ({ value: net.id, display: net.name }))
            }
          ]} />
        </Form>
      </div>
    </div>
  );
}