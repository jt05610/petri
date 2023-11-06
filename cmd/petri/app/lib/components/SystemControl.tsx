import { useContextSelector } from "use-context-selector";
import { RecordRunContext } from "~/context";
import { Suspense } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import Timeline from "~/lib/components/timeline";
import { PetriNetContext } from "~/lib/context/petrinet";
import { DeviceControl } from "~/lib/components/DeviceControl";
import { ParserProvider } from "~/lib/context/ParserContext";
import ParameterEditor from "~/lib/components/ParameterEditor";
import { parameterScope } from "~/util/parameters";

export function SystemControl() {
  const net = useContextSelector(PetriNetContext, (context) => context);
  const run = useContextSelector(RecordRunContext, (context) => context?.run);
  return (
    <Suspense fallback={<div>Loading...</div>}>
      {run && (
        <ParserProvider scope={parameterScope(run.parameters)}>
          <div className={"flex flex-col w-full items-center justify-items-center"}>
            <div className={"w-full flex-col overflow-auto"}>
              <div className={"flex h-1/4 w-full flex-col space-x-2 p-2 overflow-y-scroll"}>
                <div
                  className={"flex h-full w-full border-2 border-gray-900 rounded-xl items-center p-2 space-x-2"}
                >
                  {net && <MarkedNet marking={net.petriNet.marking} net={net.petriNet.net} />}
                </div>
              </div>
              <ParameterEditor />
              <div className={"flex h-2/4 flex-row flex-wrap gap-2"}>
                {net && net.petriNet.net.deviceEvents(net.petriNet.marking).map((
                    device, index) => {
                    return <DeviceControl {...device} key={index} />;
                  }
                )}
              </div>
            </div>
            <div className={"w-full max-w-screen"}>
              <Suspense fallback={<div>Loading...</div>}>
                {net && run &&
                  <Timeline sequence={run} />
                }
              </Suspense>
            </div>
          </div>
        </ParserProvider>
      )}
    </Suspense>

  )
    ;
}