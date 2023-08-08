import { PetriNetContext, SessionProvider } from "~/context";
import { useContextSelector } from "use-context-selector";
import { SystemControl } from "~/lib/components/systemControl";
import type { LoaderArgs } from "@remix-run/node";
import { getRunSession } from "~/models/net.run.session.server";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";
import { useLoaderData } from "@remix-run/react";
import { json } from "@remix-run/node";
import Player from "~/lib/components/player";
import { Suspense } from "react";

export const loader = async ({ params, request }: LoaderArgs) => {
  const { sessionID } = params;
  invariant(sessionID, "sessionID is required");
  await requireUser(request);
  const details = await getRunSession({ id: sessionID });
  return json({ session: details });
};

export default function PlaySequence() {
  const { session } = useLoaderData<typeof loader>();
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <div className={"h-7/10 w-full"}>
        <Suspense fallback={<div>Loading controls</div>}>
          {petriNet && <SystemControl net={petriNet!} />}
        </Suspense>
      </div>
      <div className={"h-3/10 w-full"}>
        <Suspense fallback={<div>Loading session</div>}>
          {marking && petriNet && session &&
            <SessionProvider sessionDetails={session}>
              <Player />
            </SessionProvider>
          }
        </Suspense>
      </div>
    </div>
  );
};