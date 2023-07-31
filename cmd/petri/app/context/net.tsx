import { createContext, useReducer } from "react";
import type { ReactNode, Dispatch, Reducer } from "react";
import type { Net, Arc, Place, Transition } from "@prisma/client";
import type { FormInput } from "~/lib/components/formInput";

export type NetState = {
  id: Net["id"];
  places: Pick<Place, "id" | "name">[];
  transitions: Pick<Transition, "id" | "name">[];
  arcs: Pick<Arc, "id" | "placeID" | "transitionID" | "fromPlace">[];
}

export const NetContext = createContext<NetState>({ id: "", places: [], transitions: [], arcs: []});
export const NetDispatchContext = createContext<Dispatch<NetAction> | null>(null);

type NetProviderProps = {
  value: NetState;
  children: ReactNode;
}

export function NetProvider(props: NetProviderProps) {
  const [net, dispatch] = useReducer<Reducer<NetState, NetAction>>(netReducer, props.value);
  return (
    <NetContext.Provider value={net}>
      <NetDispatchContext.Provider value={dispatch}>
        {props.children}
      </NetDispatchContext.Provider>
    </NetContext.Provider>
  );
}

export enum NetActionKind {
  AddPlace = "add-place",
  AddTransition = "add-transition",
  AddArc = "add-arc",
  DeletePlace = "delete-place",
  DeleteTransition = "delete-transition",
  DeleteArc = "delete-arc",
  UpdatePlace = "update-place",
  UpdateTransition = "update-transition",
}

interface NetAction {
  type: NetActionKind;
  payload: FormInput;
}

function netReducer(net: NetState, action: NetAction): NetState {
  switch (action.type) {
    case NetActionKind.AddPlace: {
      if (net.places.find((place) => place.id === action.payload.id)) {
        break;
      }
      net.places.push(action.payload as Place);
      break;
    }
    case NetActionKind.AddTransition: {
      if (net.transitions.find((transition) => transition.id === action.payload.id)) {
        break;
      }
      net.transitions.push(action.payload as Transition);
      break;
    }
    case NetActionKind.AddArc: {
      if (net.arcs.find((arc) => arc.id === action.payload.id)) {
        break;
      }
      net.arcs.push(action.payload as Arc);
      break;
    }
    case NetActionKind.DeletePlace: {
      return {
        ...net,
        places: net.places.filter((place) => place.id !== action.payload.id)
      };
    }
    case NetActionKind.DeleteTransition: {
      return {
        ...net,
        transitions: net.transitions.filter((transition) => transition.id !== action.payload.id)
      };
    }
    case NetActionKind.DeleteArc: {
      return {
        ...net,
        arcs: net.arcs.filter((arc) => arc.id !== action.payload.id)
      };
    }
    case NetActionKind.UpdatePlace: {
      return {
        ...net,
        places: net.places.map((place) => {
          if (place.id === action.payload.id) {
            return {
              ...place,
              ...action.payload
            };
          }
          return place;
        })
      };
    }
    case NetActionKind.UpdateTransition: {
      return {
        ...net,
        transitions: net.transitions.map((transition) => {
          if (transition.id === action.payload.id) {
            return {
              ...transition,
              ...action.payload
            };
          }
          return transition;
        })
      };
    }
    default:
      break;

  }
  return net;
}

