import { createContext } from "use-context-selector";
import { PetriNet } from "~/util/petrinet";
import type { Dispatch, ReactNode, Reducer } from "react";
import { useReducer } from "react";
import type { NetDetailsWithChildren } from "~/models/net.server";
import cloneDeep from "lodash/cloneDeep";

type Marking = {
  [key: string]: number
}

export type ActivePetriNet = {
  userID?: string
  net: PetriNet
  marking: Marking
  markingHistory: Marking[];
  enabledEvents: { [eventID: string]: boolean }
}


function initialNet(net: NetDetailsWithChildren): PetriNet {
  if (net.devices && net.devices.length > 0) {
    return new PetriNet(net);
  } else {
    return new PetriNet(net).combinedNet;
  }
}

type InitialActiveNetProps = {
  net: NetDetailsWithChildren
  userID?: string
}

const initialActiveNet = ({ net, userID }: InitialActiveNetProps): ActivePetriNet => {
  const pNet = initialNet(net);
  return {
    net: pNet,
    marking: pNet.initialMarking,
    markingHistory: [pNet.initialMarking],
    enabledEvents: pNet.allEvents(pNet.initialMarking).reduce((acc, event) => ({
      ...acc,
      [event.id]: pNet.eventEnabled(pNet.initialMarking, event.id)
    }), {})
  };
};

export const PetriNetContext = createContext<({
  petriNet: ActivePetriNet;
  dispatch: Dispatch<PetriNetAction>
}) | null>(null);

type PetriNetProviderProps = {
  net: NetDetailsWithChildren,
  userID: string,
  children: ReactNode;
}


type UpdateMarkingPayload = {
  [key: string]: number
}

type RevertMarkingPayload = {
  markingIndex: number
}

export enum PetriNetActionType {
  UpdateMarking = "updateMarking",
  RevertMarking = "revertMarking"
}

type UpdateMarkingAction = {
  type: PetriNetActionType.UpdateMarking;
  payload: UpdateMarkingPayload;
}

type RevertMarkingAction = {
  type: PetriNetActionType.RevertMarking;
  payload: RevertMarkingPayload;
}


type PetriNetAction = UpdateMarkingAction | RevertMarkingAction;

function petriNetReducer(state: ActivePetriNet, action: PetriNetAction): ActivePetriNet {
  switch (action.type) {
    case PetriNetActionType.UpdateMarking: {
      // Use cloneDeep to create a new copy of state.marking
      const newMarking = cloneDeep(action.payload);
      const markingHistory = [...state.markingHistory, newMarking];
      const enabledEvents = state.net.allEvents(newMarking).reduce((acc, event) => ({
        ...acc,
        [event.id]: state.net.eventEnabled(newMarking, event.id)
      }), {});
      console.log("marking history", markingHistory);
      return {
        ...state,
        marking: newMarking,
        markingHistory,
        enabledEvents
      };

    }
    case PetriNetActionType.RevertMarking: {
      const marking = state.markingHistory[action.payload.markingIndex];
      const markingHistory = state.markingHistory.slice(0, action.payload.markingIndex + 1);
      const enabledEvents = state.net.allEvents(marking).reduce((acc, event) => ({
        ...acc,
        [event.id]: state.net.eventEnabled(marking, event.id)
      }), {});
      return {
        ...state,
        marking,
        markingHistory,
        enabledEvents
      };
    }
    default: {
      return state;
    }
  }
}


export function PetriNetProvider({ net, userID, children }: PetriNetProviderProps) {
  const [petriNet, dispatch] = useReducer<Reducer<ActivePetriNet, PetriNetAction>>(petriNetReducer,
    initialActiveNet({ net }));

  return (
    <PetriNetContext.Provider value={{ petriNet, dispatch }}>
      {children}
    </PetriNetContext.Provider>
  );
}

