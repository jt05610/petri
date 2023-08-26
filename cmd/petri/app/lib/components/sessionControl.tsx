import { useContextSelector } from "use-context-selector";
import { RunSessionContext, SessionActionType } from "~/context";
import { MarkedNet } from "~/lib/components/markedNet";
import type { Event } from "@prisma/client";
import { PetriNetActionType, PetriNetContext } from "~/lib/context/petrinet";


export function SessionControl() {
  const net = useContextSelector(PetriNetContext, (context) => context?.petriNet);
  const netDispatch = useContextSelector(PetriNetContext, (context) => context?.dispatch);
  const session = useContextSelector(RunSessionContext, (context) => context?.session);
  const dispatch = useContextSelector(RunSessionContext, (context) => context?.dispatch);

  /*
    socket?.on("event", (data: EventReceived) => {
      if (!dispatch) return;
      if (!session || !session.run || !session.run.steps) {
        return;
      }
      if (!session.activeAction) {
        return;
      }
      if (!playing) {
        return;
      }
      dispatch({ type: SessionActionType.ActionCompleted, payload: data });
      dispatch({ type: SessionActionType.ActionStarted, payload: {} });
      handleEvent(session.activeAction.action.event, session.activeAction.action.deviceId, data);
    });
  */

  function handleEvent(event: Pick<Event, "id" | "name">, deviceID: string, data: any) {
    console.log("handle event", event, deviceID, data);
    if (!net || !netDispatch) return;
    const newMarking = net.net.handleEvent(net.marking, event.id);
    //const instance = selectedInstances[deviceID];
    //socket?.emit("command", { data, input: marking, output: newMarking, deviceID: instance, command: event.name.replace(/\s/g, "_").toLowerCase() } as Command);
    netDispatch({ type: PetriNetActionType.UpdateMarking, payload: newMarking });
  }

  const handleStartActionClicked = () => {
    const step = session?.activeAction;
    console.log("start action", step);
    if (!step) return;
    handleEvent(step.action.event, step.action.deviceId, session?.data);
  };

  const handleStopActionClicked = () => {
    const step = session?.activeAction;
    console.log("stop action");
    if (!step) return;
    console.log(step);
    handleEvent(step.action.event, step.action.deviceId, session?.data);
  };

  const handleInitializeClicked = () => {
    console.log("initialize");
    if (!dispatch) return;
    dispatch({ type: SessionActionType.StartSession, payload: {} });
  };

  return (
    <div className={"flex w-full flex-col space-y-2"}>
      {net && <MarkedNet marking={net.marking} net={net.net} />}
      <div className={"flex flex-row space-x-2"}>
        <button className={"rounded-full p-2"} onClick={handleInitializeClicked}>
          Initialize
        </button>
        <button className={"rounded-full p-2 bg-green-700"} onClick={handleStartActionClicked}>
          Start
        </button>
        <button className={"rounded-full p-2 bg-red-700"} onClick={handleStopActionClicked}>
          Stop
        </button>
      </div>
    </div>
  );
}