import type { ActionArgs } from "@remix-run/node";
import { createSession } from "~/models/net.run.session.server";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { redirect } from "@remix-run/node";
import { Form } from "@remix-run/react";

export const action = async ({ params, request }: ActionArgs) => {
  const { sequenceID } = params;
  invariant(sequenceID, "sequenceID is required");
  const user = await requireUser(request);
  const session = await createSession({
    runID: sequenceID,
    userID: user.id
  });
  return redirect(session.id);
};

export default function PlaySequence() {
  return (
    <Form method={"post"}>
      <button type={"submit"}>
        New session
      </button>
    </Form>
  );
};