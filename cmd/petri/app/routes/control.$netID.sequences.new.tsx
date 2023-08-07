import type { LoaderArgs } from "@remix-run/node";
import { redirect } from "@remix-run/node";
import { addSequence, SequenceInputSchema } from "~/models/sequence.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import invariant from "tiny-invariant";

export const action = async ({ params, request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const user = await getUserById(userID);
  if (!user) {
    throw new Error("User not found");
  }
  invariant(params.netID, "netID is required");
  let formData = await request.json();
  formData["netID"] = params.netID;
  const submission = SequenceInputSchema.parse(formData);
  const sequence = await addSequence(submission);
  return redirect(`/control/${params.netID}/sequences/${sequence.id}`);
};
