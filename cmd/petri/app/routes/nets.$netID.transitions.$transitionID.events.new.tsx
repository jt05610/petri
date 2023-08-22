import type { LoaderArgs } from "@remix-run/node";
import { redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import { addEvent, EventFieldSchema, EventInputSchema } from "~/models/net.transition.event.server";
import { Form, useActionData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import { Suspense, useState } from "react";
import  Modal  from "~/lib/components/modal";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.eventID, "eventID not found");
  let formData = await request.formData();
  const submission = parse(formData, {
    schema: EventInputSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  const newEvent = await addEvent(submission.value);
  return redirect(`/nets/${params.netID}/events/${newEvent.id}`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const authorID = await requireUserId(request);
  const user = await getUserById(authorID);
  if (!user) {
    throw new Error("User not found");
  }
  return {};
};


export default function Event() {
  const [isOpen, setIsOpen] = useState(true); // handle modal open state
  const lastSubmission = useActionData<typeof action>();
  const [form, { name, description, fields }] = useForm({
    lastSubmission,
    onValidate({ formData }) {
      return parse(formData, { schema: EventInputSchema });
    }
  });

  if (!isOpen) return null; // if not open, do not render modal and content

  return (
    // if not open then we want to show a button to reopen the modal
    <Suspense fallback={<div>Loading...</div>}>
      {isOpen ? (

        // if open then we want to show the modal
        <Modal isOpen={isOpen} setIsOpen={setIsOpen} name={"New event"}>
          <Form method={"post"} {...form.props}>
            <FormContent fields={[
              {
                type: "text",
                name: "name",
                content: name.form,
                error: name.error
              },
              {
                name: description.name,
                type: "textarea",
                content: description ? description.form : "",
                error: description.error
              },
              {
                name: fields.name,
                type: "array",
                content: fields.form,
                error: fields.error,
                arraySchema: EventFieldSchema,
                arrayFields: [
                  {
                    name: "name",
                    type: "text",
                    content: fields.form,
                    error: fields.error
                  },
                  {
                    name: "type",
                    type: "select",
                    content: fields.form,
                    error: fields.error,
                    options: [
                      { value: "string", display: "String" },
                      { value: "number", display: "Number" },
                      { value: "boolean", display: "Boolean" },
                      { value: "date", display: "Date" },
                      { value: "time", display: "Time" },
                      ],
                  },
                  {
                    name: "description",
                    type: "textarea",
                    content: fields.form,
                    error: fields.error
                  },
                ]
              }
            ]} />
          </Form>
        </Modal>

      ) :(
        <button
          type="button"
          onClick={() => setIsOpen(true)}
        >
          Open Modal
        </button>
        )}

    </Suspense>


  );
}