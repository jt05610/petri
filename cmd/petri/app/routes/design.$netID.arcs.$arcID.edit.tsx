import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getArc, updateArc, ArcUpdateFormSchema } from "~/models/net.arc.server";
import { Form, useActionData, useLoaderData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { useState } from "react";
import { listTransitions } from "~/models/net.transition.server";
import { listPlaces } from "~/models/net.place.server";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.arcID, "arcID not found");
  let formData = await request.formData();
  formData.append("id", params.arcID);
  const submission = parse(formData, {
    schema: ArcUpdateFormSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  await updateArc(submission.value);
  return redirect(`/design/${params.netID}/arcs/${params.arcID}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.netID, "netID not found");
  invariant(params.arcID, "netID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const arc = await getArc({ id: params.arcID });

  const places = await listPlaces({ netID: params.netID });
  const transitions = await listTransitions({ netID: params.netID });
  return json({ arc: arc, places, transitions });
};

export default function Arc() {
  const lastSubmission = useActionData<typeof action>();
  const { arc, places, transitions } = useLoaderData<typeof loader>();
  const [changed, setChanged] = useState(false);
  const [form, { fromPlace, placeID, transitionID }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      setChanged(false);
      return parse(formData, { schema: ArcUpdateFormSchema });
    }
  });

  return (
    <div className={"rounded-lg border-2 p-2 "}>
      <h2 className={"text-lg font-semibold"}>Update</h2>
      <Form method={"post"} {...form.props} onChange={() => setChanged(true)}>
        <FormContent activeButton={changed} fields={[
          {
            name: placeID.name,
            type: "select",
            content: arc.place.id,
            error: placeID.error,
            options: places.map((place) => ({ display: place.name, value: place.id }))
          },
          {
            name: transitionID.name,
            type: "select",
            content: arc.transition.id,
            error: transitionID.error,
            options: transitions.map((transition) => ({ display: transition.name, value: transition.id }))
          },
          {
            type: "checkbox",
            name: fromPlace.name,
            content: arc.fromPlace.toString(),
            error: fromPlace.error
          }
        ]} />
      </Form>
    </div>
  );
}