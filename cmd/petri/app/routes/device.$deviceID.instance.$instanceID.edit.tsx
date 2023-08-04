import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { addInstance, getInstance, InstanceInputSchema } from "~/models/instance.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/components/form";
import { Language } from "@prisma/client";

export const action = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  invariant(user, "User not found");
  invariant(params.deviceID, "deviceID not found");
  let formData = await request.formData();
  formData.append("authorID", authorID)
  formData.append("deviceID", params.deviceID);
  const submission = parse(formData, {
    schema: InstanceInputSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }
  const instance = await addInstance(submission.value);
  return redirect(`/device/${params.deviceID}/instance/${instance.id}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.instanceID, "deviceID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const instance = await getInstance({ id: params.instanceID });
  return json({instance})
};

export default function EditInstance() {
  const { instance } = useLoaderData<typeof loader>();
  const lastSubmission = useActionData<typeof action>();
  const [form, { name, addr, language }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      return parse(formData, { schema: InstanceInputSchema });
    }
  });
  const langOptions: { display: string, value: string }[] = [];
  for (const key in Language) {
    if (isNaN(Number(key))) {
      langOptions.push({ value: key, display: key.toLowerCase() });
    }
  }

  return (
    <div className={"flex flex-col w-full justify-center space-y-2 p-2"}>
      <h1 className={"text-2xl font-bold"}>Edit instance</h1>
      <br />
      <div className={"rounded-lg border-2 p-2 "}>
        <h2 className={"text-lg font-semibold"}>Update</h2>
        <Form method={"post"} {...form.props}>
          <FormContent fields={[
            {
              type: "text",
              name: "name",
              content: instance.name,
              error: name.error
            },
            {
              name: addr.name,
              type: "text",
              content: instance.addr,
              error: addr.error
            },
            {
              name: language.name,
              type: "select",
              content: instance.language,
              error: language.error,
              options: langOptions
            }
          ]} />
        </Form>
      </div>
    </div>
  );
}