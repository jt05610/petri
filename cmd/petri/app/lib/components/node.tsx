import type {UpdatePlaceInput} from "~/models/place.server";
import type {UpdateTransitionInput} from "~/models/transition.server";

export type NodeUpdate = (UpdatePlaceInput | UpdateTransitionInput) & {
  arcs: {
    id: string,
    other: string,
    otherKind: NodeKind,
    fromPlace: boolean
  }[]
};

export enum NodeKind {
  Place = "Place",
  Transition = "Transition"
}
