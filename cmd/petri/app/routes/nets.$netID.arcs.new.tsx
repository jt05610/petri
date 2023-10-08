import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import {  ArcInputFormSchema, addArc } from "~/models/net.arc.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { listPlaces } from "~/models/net.place.server";
import { listTransitions } from "~/models/net.transition.server";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.netID, "netID not found");
  let formData = await request.formData();
  formData.append("netID", params.netID);
  console.log(formData.get("fromPlace"))
  const submission = parse(formData, {
    schema: ArcInputFormSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }
  const arc = await addArc(submission.value);
  return redirect(`/nets/${params.netID}/arcs/${arc.id}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const places = await listPlaces({ netID: params.netID });
  const transitions = await listTransitions({ netID: params.netID });
  return json({ places, transitions });
};

export default function Arc() {
  const lastSubmission = useActionData<typeof action>();
  const { places, transitions } = useLoaderData<typeof loader>();
  const [form, { fromPlace, placeID, transitionID }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      return parse(formData, { schema: ArcInputFormSchema });
    }
  });

  return (
    <div className={"rounded-lg border-2 p-2 "}>
      <h2 className={"text-lg font-semibold"}>Update</h2>
      <Form method={"post"} {...form.props} >
        <FormContent fields={[
          {
            name: placeID.name,
            type: "select",
            content: lastSubmission?.payload?.placeID,
            error: placeID.error,
            options: places.map((place) => ({ display: place.name, value: place.id }))
          },
          {
            name: transitionID.name,
            type: "select",
            content: lastSubmission?.payload?.transitionID,
            error: transitionID.error,
            options: transitions.map((transition) => ({ display: transition.name, value: transition.id }))
          },
          {
            type: "checkbox",
            name: fromPlace.name,
            content: lastSubmission?.payload?.fromPlace,
            error: fromPlace.error,
          }
        ]} />
      </Form>
    </div>
  );

}