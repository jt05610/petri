import type { LoaderArgs } from "@remix-run/node";
import { redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import type { EventInput} from "~/models/net.transition.event.server";
import { addEvent, EventFieldSchema, EventInputSchema } from "~/models/net.transition.event.server";
import { Form, useActionData, useSubmit } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useFieldList, useForm } from "@conform-to/react";
import FormContent from "~/lib/layouts/form";
import type { FormEvent } from "react";
import { Suspense, useState } from "react";
import Modal from "~/lib/components/modal";

export const action = async ({ params, request }: LoaderArgs) => {
  invariant(params.transitionID, "transitionID not found");
  let formData = await request.formData();
  formData.set("transitionID", params.transitionID);
  const submission = parse(formData, {
    schema: EventInputSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  const newEvent = await addEvent(submission.value);
  return redirect(`/nets/${params.netID}/events/${newEvent.id}`);
};

export const loader = async ({ request }: LoaderArgs) => {
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
  const [form, { name, description, fields }] = useForm<EventInput>({
    lastSubmission,
    onValidate({formData}) {
      return parse(formData, { schema: EventInputSchema });
    },
    shouldValidate: 'onBlur',
  });
  const fieldsList = useFieldList(form.ref, fields);
  const submit = useSubmit();

  if (!isOpen) return null; // if not open, do not render modal and content

  return (
    // if not open then we want to show a button to reopen the modal
    <Suspense fallback={<div>Loading...</div>}>
      {isOpen ? (
        // if open then we want to show the modal
        <Modal isOpen={isOpen} setIsOpen={setIsOpen} name={"New event"}>
          <Form className={`flex flex-col`} method={"post"} {...form.props} onSubmit={(e: FormEvent) => {
            const formData = new FormData(e.target as HTMLFormElement);
            console.log(formData);
            submit(formData, { method: "post", replace: true });
          }
          }>
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
                arrayFields: fieldsList,
              }
            ]} />
          </Form>
        </Modal>

      ) : (
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