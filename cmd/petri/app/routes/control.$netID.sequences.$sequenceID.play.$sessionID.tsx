import { PetriNetContext, SessionProvider } from "~/context";
import { useContextSelector } from "use-context-selector";
import Timeline from "~/lib/components/timeline";
import { SystemControl } from "~/lib/components/systemControl";
import type { LoaderArgs } from "@remix-run/node";
import { getRunSession } from "~/models/net.run.session.server";
import invariant from "tiny-invariant";
import { requireUser } from "~/session.server";


export const loader = async ({ params, request }: LoaderArgs) => {
  const { sessionID } = params;
  invariant(sessionID, "sessionID is required");
  await requireUser(request);
  return getRunSession({ id: sessionID });
};

export default function PlaySequence() {
  const session = useContextSelector(PetriNetContext, (context) => context?.session);
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  console.log(petriNet?.toGraphViz());
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <div className={"h-7/10 w-full"}>
        <SystemControl net={petriNet!} />
      </div>
      <div className={"h-3/10 w-full"}>
        {marking && petriNet &&
          <SessionProvider sessionDetails={session}>
            <Timeline />
          </SessionProvider>
        }
      </div>
    </div>
  );
};