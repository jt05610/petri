import type { UpdatePlaceInput, UpdateTransitionInput } from "~/models/net.server";

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

export type Node = {
  id: string,
  name: string,
  kind: NodeKind,
  data: NodeUpdate,
}
