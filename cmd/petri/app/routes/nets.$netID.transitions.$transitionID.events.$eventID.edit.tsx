import type { LoaderArgs } from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { getEvent, updateEvent, UpdateEventSchema } from "~/models/net.transition.event.server";
import { Form, useActionData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm, useFieldList, list } from "@conform-to/react";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.eventID, "eventID not found");
  let formData = await request.formData();
  formData.append("id", params.eventID);
  const submission = parse(formData, {
    schema: UpdateEventSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  await updateEvent(submission.value);
  return redirect(`/design/${params.transitionID}/events/${params.eventID}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  invariant(params.eventID, "transitionID not found");
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  const event = await getEvent({ id: params.eventID });
  return json({ event: event });
};

export default function Event() {
  const lastSubmission = useActionData<typeof action>();
  const [form, { name, description, fields }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      return parse(formData, { schema: UpdateEventSchema });
    }
  });
  const fieldList = useFieldList(form.ref, fields);

  return (
    <div className={"rounded-lg border-2 p-2 "}>
      <h2 className={"text-lg font-semibold"}>Update</h2>
      <Form method={"post"} {...form.props}>
        <label className={"block"}>
          <span className={"text-gray-700"}>Name</span>
          <input className={"form-input mt-1 block w-full"} name={name.name} />
        </label>
        <label className={"block"}>
          <span className={"text-gray-700"}>Description</span>
          <textarea className={"form-input mt-1 block w-full"} name={description.name} />
        </label>
        <label className={"block"}>
          Fields
        </label>
        <div className={"space-y-2"}>
          {fieldList.map((field, index) => (
            <div key={field.key} className={"space-y-2"}>
              <label className={"block"}>
                <span className={"text-gray-700"}>Name</span>
                <input className={"form-input mt-1 block w-full"} name={"name"} />
              </label>
              <label className={"block"}>
                <span className={"text-gray-700"}></span>
                <textarea className={"form-input mt-1 block w-full"} name={"description"} />
              </label>
              <label className={"block"}>
                <span className={"text-gray-700"}>Type</span>
                <select className={"form-input mt-1 block w-full"} name={"type"}>
                  <option value={"string"}>String</option>
                  <option value={"number"}>Number</option>
                  <option value={"boolean"}>Boolean</option>
                </select>
              </label>
              <span>{field.error}</span>
              <button className={"btn rounded-full px-2 py-1 bg-red-500"} {...list.remove(fields.name, {index})}>Remove</button>
            </div>
          ))}
          <button className={"btn rounded-full px-2 py-1 bg-teal-500"} {...list.append(fields.name)}> Add Field</button>
        </div>
      </Form>
    </div>
  );

}