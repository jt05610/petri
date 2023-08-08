import type { PetriNet } from "~/util/petrinet";
import { useContextSelector } from "use-context-selector";
import { PetriNetContext, RunSessionContext, SessionActionType, SocketContext } from "~/context";
import { useState } from "react";
import { MarkedNet } from "~/lib/components/markedNet";
import type { Event } from "@prisma/client";
import type { Command, EventReceived } from "../../../server/command";

type SystemControlProps = {
  net: PetriNet
}

export function SessionControl(props: SystemControlProps) {
  const marking = useContextSelector(PetriNetContext, (context) => context?.marking);
  const setMarking = useContextSelector(PetriNetContext, (context) => context?.setMarking);
  const socket = useContextSelector(SocketContext, (context) => context);
  const session = useContextSelector(RunSessionContext, (context) => context?.session);
  const dispatch = useContextSelector(RunSessionContext, (context) => context?.dispatch);
  const [playing, setPlaying] = useState(false);

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

  const [selectedInstances, setSelectedInstances] = useState<{
    [deviceID: string]: string
  }>({} as {
    [deviceID: string]: string
  });

  const handleSelectChanged = (deviceID: string, instanceID: string) => {
    setSelectedInstances({ ...selectedInstances, [deviceID]: instanceID });
  };

  function handleEvent(event: Pick<Event, "id" | "name">, deviceID: string, data: any) {
    if (!selectedInstances[deviceID]) {
      console.log("no instance selected");
      return;
    }
    console.log("handle event", event, deviceID, data);
    if (!props.net || !marking || !setMarking) return;
    const newMarking = props.net.handleEvent(marking, event.id);
    const instance = selectedInstances[deviceID];
    socket?.emit("command", { data, input: marking, output: newMarking, deviceID: instance, command: event.name.replace(/\s/g, "_").toLowerCase() } as Command);
    setMarking(newMarking);
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
    setPlaying(false);
    handleEvent(step.action.event, step.action.deviceId, session?.data);
  };

  const handleInitializeClicked = () => {
    console.log("initialize");
    if (!dispatch) return;
    dispatch({ type: SessionActionType.StartSession, payload: {} });
  };

  return (
    <div className={"flex w-full flex-col space-y-2"}>
      {marking && <MarkedNet marking={marking} net={props.net} />}
      <div className={"flex flex-row space-x-2"}>
        {props.net.childDeviceEvents(marking!).map(({ id, name, instances, events }) => {
          return (
            <div
              key={id}
              className={"flex flex-col space-y-2 border-md rounded-lg p-2"}
            >
              <h2 className={"text-xl font-bold"}>{name}</h2>
              <select
                defaultValue={""}
                className={"rounded-full p-2"}
                onChange={(e) => {
                  const instance = instances.find((instance) => instance.id === e.target.value);
                  if (!instance) return;
                  handleSelectChanged(id, instance.id);
                  const data = { data: {}, deviceID: instance.id, command: "get" };
                  console.log("systemControl sent", data);
                  socket?.emit("state", data);
                }
                }>
                <option value={""}>Select a device</option>
                {instances?.map((instance) => {
                    return (
                      <option key={instance.id} value={instance.id}>{instance.name}</option>
                    );
                  }
                )}
              </select>
            </div>
          );
        })}
      </div>
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