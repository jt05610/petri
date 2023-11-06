import type { LoaderArgs } from "@remix-run/node";
import { redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import invariant from "tiny-invariant";
import { getUserById } from "~/models/user.server";
import type { EventInput } from "~/models/net.transition.event";
import { EventInputSchema } from "~/models/net.transition.event";
import { addEvent } from "~/models/net.transition.event.server";
import { Form, useActionData } from "@remix-run/react";
import { parse } from "@conform-to/zod";
import { badRequest } from "~/util/request.server";
import { useFieldList, useForm } from "@conform-to/react";
import { Suspense, useState } from "react";
import Modal from "~/lib/components/modal";
import { Field, FieldSetTable, FieldTextArea } from "~/lib/components/FormFieldSet";

export const action = async ({ params, request }: LoaderArgs) => {
  console.log("params", params);
  const transitionID = params.transitionID;
  invariant(transitionID, "transitionID not found");
  let formData = await request.formData();
  formData.set("transitionID", transitionID);
  const submission = parse(formData, {
    schema: EventInputSchema
  });
  if (!submission.value || submission.intent !== "submit") {
    return badRequest(submission);
  }

  const newEvent = await addEvent(transitionID, submission.value);
  return redirect(`/nets/${params.netID}/transitions/${transitionID}/events/${newEvent.id}`);
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
    onValidate({ formData }) {
      const sub = parse(formData, { schema: EventInputSchema });
      console.log("sub", sub);
      return sub;
    },
    shouldValidate: "onBlur"
  });
  const fieldsList = useFieldList(form.ref, fields);

  if (!isOpen) return null; // if not open, do not render modal and content

  return (
    // if not open then we want to show a button to reopen the modal
    <Suspense fallback={<div>Loading...</div>}>
      {isOpen ? (
        // if open then we want to show the modal
        <Modal isOpen={isOpen} setIsOpen={setIsOpen} name={"New event"}>
          <Form className={`flex flex-col`} method={"post"} {...form.props}>
            <Field {...name} />
            <FieldTextArea {...description} />
            <h2 className={`text-xl font-bold`}>Fields</h2>
            <FieldSetTable config={fields} fields={[
              "name",
              {
                name: "type",
                options: ["string", "number", "boolean",]
              },
              "description",
              "min",
              "max",
              "choices",
              "unit"
            ]}
                           fieldList={fieldsList} defaultValue={{
              name: "event",
              description: "event description",
              type: "string",
              choices: []
            }} />
            <button
              type="submit"
              className={`bg-blue-900 text-white rounded-full p-2`}
            >
              Submit
            </button>
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