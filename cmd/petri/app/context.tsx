import type { Dispatch, ReactNode, Reducer } from "react";
import { createContext, useContextSelector } from "use-context-selector";
import type { Socket } from "socket.io-client";
import { PetriNet } from "~/util/petrinet";
import { useReducer, useState } from "react";
import type { NetDetailsWithChildren } from "~/models/net.server";
import type { Place } from "@prisma/client";
import type {
  DeviceInputDisplay,
  SequenceInputDisplay,
  SequenceEventInputDisplay
} from "~/models/sequence.server";


type ProviderProps = {
  socket: Socket | undefined;
  children: ReactNode;
};

export const SocketContext = createContext<Socket | undefined>(undefined);

export function SocketProvider({ socket, children }: ProviderProps) {
  return <SocketContext.Provider value={socket}>{children}</SocketContext.Provider>;
}

export const PetriNetContext = createContext<({
  petriNet: PetriNet;
  marking: { [key: string]: number };
  setMarking: (marking: { [key: string]: number }) => void;
}) | null>(null);

type PetriNetProviderProps = {
  net: NetDetailsWithChildren;
  children: ReactNode;
}
const netMarking = (places: Pick<Place, "id">[], initial: number[]) => {
  return places.map((p, i) => [p.id, initial[i]]).reduce((acc, [id, tokens]) => ({ ...acc, [id]: tokens }), {});
};

export function PetriNetProvider({ net, children }: PetriNetProviderProps) {
  const [petriNet] = useState<PetriNet>(net.device?.instances ? new PetriNet(net) : new PetriNet(net).combinedNet);
  const [marking, setMarking] = useState<{ [key: string]: number }>(netMarking(petriNet.net.places, petriNet.net.initialMarking));
  return (
    <PetriNetContext.Provider value={{ petriNet, marking, setMarking }}>
      {children}
    </PetriNetContext.Provider>
  );
}

export const RecordSequenceContext = createContext<({
  sequence: SequenceInputDisplay;
  dispatch: Dispatch<SequenceAction>;
}) | null>(null);

type RemoveDeviceEventPayload = {
  sequenceEventIndex: number;
  deviceEventIndex: number;
}

type DeviceEventMovedPayload = {
  sequenceEventIndex: number;
  deviceEventIndex: number;
  newSequenceEventIndex: number;
}

type SequenceEventNoteAdded = {
  sequenceEventIndex: number;
  note: string;
}

type SequenceEventNoteRemoved = {
  sequenceEventIndex: number;
  noteIndex: number;
}

type DeviceEventChanged = {
  sequenceEventIndex: number;
  deviceEventIndex: number;
  deviceEvent: DeviceInputDisplay;
}

type ReducerActionPayload =
  DeviceInputDisplay
  | SequenceEventInputDisplay
  | RemoveDeviceEventPayload
  | DeviceEventMovedPayload
  | SequenceEventNoteAdded
  | SequenceEventNoteRemoved
  | DeviceEventChanged

export enum SequenceActionType {
  EVENT_ADDED = "eventAdded",
  EVENT_REMOVED = "eventRemoved",
  DEVICE_EVENT_MOVED = "deviceEventMoved",
  NOTE_ADDED_TO_EVENT = "noteAddedToEvent",
  NOTE_REMOVED_FROM_EVENT = "noteRemovedFromEvent",
  DEVICE_EVENT_CHANGED = "deviceEventChanged",
}

type SequenceAction = {
  type: SequenceActionType;
  payload: ReducerActionPayload;
}

const defaultInitial: SequenceInputDisplay = {
  deviceNames: [],
  description: "",
  name: "",
  netID: "",
  events: []
};


type RecordSequenceProviderProps = {
  initialSequence?: SequenceInputDisplay;
  children: ReactNode;
}

export const addDeviceNames = (sequence: SequenceInputDisplay, pNet: PetriNet): SequenceInputDisplay => {
  return {
    ...sequence,
    deviceNames: pNet.devices.map((d) => d.name)
  };
};

export function RecordSequenceProvider({ initialSequence, children }: RecordSequenceProviderProps) {
  const [sequence, dispatch] = useReducer<Reducer<SequenceInputDisplay, SequenceAction>>(sequenceReducer, initialSequence || defaultInitial);
  return (
    <RecordSequenceContext.Provider value={{ sequence, dispatch }}>
      {children}
    </RecordSequenceContext.Provider>
  );
}

function sequenceReducer(state: SequenceInputDisplay, action: { type: string; payload: ReducerActionPayload }): SequenceInputDisplay {
  switch (action.type) {
    case SequenceActionType.EVENT_ADDED: {
      const devInput = action.payload as DeviceInputDisplay;
      const events = [...state.events];
      const newSequenceEvent: SequenceEventInputDisplay = {
        sequenceID: "",
        events: [devInput],
        name: `Step ${events.length + 1}`,
        notes: []
      };
      events.push(newSequenceEvent);
      return { ...state, events };
    }
    case SequenceActionType.EVENT_REMOVED: {
      const { sequenceEventIndex, deviceEventIndex } = action.payload as RemoveDeviceEventPayload;
      const events = [...state.events];
      const sequenceEvent = events[sequenceEventIndex];
      const deviceEvents = [...sequenceEvent.events];
      deviceEvents.splice(deviceEventIndex, 1);
      if (deviceEvents.length === 0) {
        events.splice(sequenceEventIndex, 1);
      }
      return { ...state, events };
    }
    case SequenceActionType.DEVICE_EVENT_MOVED: {
      const { sequenceEventIndex, deviceEventIndex, newSequenceEventIndex } = action.payload as DeviceEventMovedPayload;
      const events = [...state.events];
      const sequenceEvent = events[sequenceEventIndex];
      const deviceEvents = [...sequenceEvent.events];
      const deviceEvent = deviceEvents.splice(deviceEventIndex, 1)[0];
      const newSequenceEvent = events[newSequenceEventIndex];
      newSequenceEvent.events.push(deviceEvent);
      if (deviceEvents.length === 0) {
        events.splice(sequenceEventIndex, 1);
      }
      return { ...state, events };
    }
    case SequenceActionType.NOTE_ADDED_TO_EVENT: {
      const { sequenceEventIndex, note } = action.payload as SequenceEventNoteAdded;
      const events = [...state.events];
      const sequenceEvent = events[sequenceEventIndex];
      sequenceEvent.notes.push(note);
      return { ...state, events };
    }
    case SequenceActionType.NOTE_REMOVED_FROM_EVENT: {
      const { sequenceEventIndex, noteIndex } = action.payload as SequenceEventNoteRemoved;
      const events = [...state.events];
      const sequenceEvent = events[sequenceEventIndex];
      sequenceEvent.notes.splice(noteIndex, 1);
      return { ...state, events };
    }
    case SequenceActionType.DEVICE_EVENT_CHANGED: {
      const { sequenceEventIndex, deviceEventIndex, deviceEvent } = action.payload as DeviceEventChanged;
      const events = [...state.events];
      const sequenceEvent = events[sequenceEventIndex];
      sequenceEvent.events[deviceEventIndex] = deviceEvent;
      return { ...state, events };
    }
    default:
      throw Error("Invalid action type");
  }

}
