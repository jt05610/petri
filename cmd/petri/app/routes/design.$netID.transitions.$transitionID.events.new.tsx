import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getTransition, updateTransition, UpdateTransitionSchema } from "~/models/transition.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/components/form";
import { useState } from "react";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.transitionID, "transitionID not found");
  let formData = await request.formData();
  formData.append("id", params.transitionID);
  const submission = parse(formData, {
    schema: UpdateTransitionSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  await updateTransition(submission.value);
  return redirect(`/design/${params.netID}/transitions/${params.transitionID}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.transitionID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const transition = await getTransition({ id: params.transitionID });
  return json({ transition: transition });
};

export default function Transition() {
  const lastSubmission = useActionData<typeof action>();
  const { transition } = useLoaderData<typeof loader>();
  const [changed, setChanged] = useState(false);
  const [form, { name, description }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      setChanged(false);
      return parse(formData, { schema: UpdateTransitionSchema });
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
              name: "name",
              content: transition.name,
              error: name.error
            },
            {
              name: description.name,
              type: "textarea",
              content: transition.description ? transition.description : "",
              error: description.error
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
          >Add Event</button>
        </Form>
      </div>
    </div>
  );

}