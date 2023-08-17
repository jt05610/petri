import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { addPlace, PlaceFormSchema } from "~/models/net.place.server";
import { Form, useActionData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";

export const action = async ({ params, request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  invariant(user, "User not found");
  invariant(params.netID, "netID not found");
  let formData = await request.formData();
  formData.append("netID", params.netID);
  const submission = parse(formData, {
    schema: PlaceFormSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }
  const place = await addPlace(submission.value);
  return redirect(`/nets/${params.netID}/places/${place.id}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  return json({});
};

export default function NewPlace() {
  const lastSubmission = useActionData<typeof action>();
  const [form, { name, description, bound }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      return parse(formData, { schema: PlaceFormSchema });
    }
  });

  return (
    <Form method={"post"} {...form.props}>
      <FormContent fields={[
        {
          type: "text",
          name: "name",
          content: name.form,
          error: name.error
        },
        {
          name: bound.name,
          type: "text",
          content: bound.form,
          error: bound.error
        },
        {
          name: description.name,
          type: "textarea",
          content: description.form,
          error: description.error
        }
      ]} />
    </Form>
  );
}