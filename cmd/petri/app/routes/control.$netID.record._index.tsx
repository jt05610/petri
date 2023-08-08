import { PetriNetContext, RecordRunProvider } from "~/context";
import { useContextSelector } from "use-context-selector";
import Timeline from "~/lib/components/timeline";
import { SystemControl } from "~/lib/components/systemControl";

export default function ControlIndex() {
  const petriNet = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  return (
    <div className={"flex flex-col h-screen w-full items-center justify-items-center"}>
      <div className={"h-7/10 w-full"}>
        <SystemControl net={petriNet!} />
      </div>
      <div className={"h-3/10 w-full"}>
        {marking && petriNet &&
          <RecordRunProvider>
            <Timeline />
          </RecordRunProvider>
        }
      </div>
    </div>
  );
};