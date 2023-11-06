import { RecordRunProvider } from "~/context";
import { SystemControl } from "~/lib/components/SystemControl";
import type { LoaderArgs} from "@remix-run/node";
import { json, redirect } from "@remix-run/node";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import invariant from "tiny-invariant";
import { AddStepsToRunSchema } from "~/models/net.run";
import { addRunSteps, getRunDetails } from "~/models/net.run.server";
import { useLoaderData } from "@remix-run/react";

export const action = async ({ params, request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  const runID = params.sequenceID;
  invariant(runID, "netID is required");
  const formData = await request.json();

  const submission = AddStepsToRunSchema.safeParse(formData);
  if (!submission.success) {
    return json(submission);
  }
  const sequence = await addRunSteps(runID, submission.data);
  return redirect(`/control/${params.netID}/sequences/${sequence.id}/`);
};

export const loader = async ({ params, request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  const netID = params.netID;
  invariant(netID, "netID is required");
  const runID = params.sequenceID;
  invariant(runID, "runID is required");
  const run = await getRunDetails({ runID });
  return json({ netID, run });
};

export default function RecordRun() {
  const { netID, run } = useLoaderData<typeof loader>();
  return (
    <RecordRunProvider netID={netID} initialRun={run}>
      <div className={"max-w-screen"}>
        <SystemControl />
      </div>
    </RecordRunProvider>
  );
}
;