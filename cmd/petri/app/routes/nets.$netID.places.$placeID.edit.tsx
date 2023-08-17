import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getPlace, updatePlace, UpdatePlaceFormSchema } from "~/models/net.place.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { useState } from "react";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.placeID, "placeID not found");
  let formData = await request.formData();
  formData.append("id", params.placeID);
  const submission = parse(formData, {
    schema: UpdatePlaceFormSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  await updatePlace(submission.value);
  return redirect(`/design/${params.netID}/places/${params.placeID}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.placeID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const place = await getPlace({ id: params.placeID });
  return json({ place: place });
};

export default function Place() {
  const lastSubmission = useActionData<typeof action>();
  const { place } = useLoaderData<typeof loader>();
  const [changed, setChanged] = useState(false);
  const [form, { name, bound, description }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      setChanged(false);
      return parse(formData, { schema: UpdatePlaceFormSchema });
    }
  });

  return (
    <div className={"rounded-lg border-2 p-2 "}>
      <h2 className={"text-lg font-semibold"}>Update</h2>
      <Form method={"post"} {...form.props} onChange={() => setChanged(true)}>
        <FormContent activeButton={changed} fields={[
          {
            type: "text",
            name: "name",
            content: place.name,
            error: name.error
          },
          {
            name: bound.name,
            type: "text",
            content: place.bound.toString(),
            error: bound.error
          },
          {
            name: description.name,
            type: "textarea",
            content: place.description ? place.description : "",
            error: description.error
          }
        ]} />
      </Form>
    </div>
  );

}