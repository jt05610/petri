import type { JsonSchema, UISchemaElement } from "@jsonforms/core";
import { addPlace, deletePlace, updatePlace } from "~/models/place.server";
import { addTransition, deleteTransition, updateTransition } from "~/models/transition.server";
import { addArc, deleteArc } from "~/models/arc.server";

import ArcInputSchema from "~/forms/ArcInput.schema.json";
import PlaceInputSchema from "~/forms/PlaceInput.schema.json";
import TransitionInputSchema from "~/forms/TransitionInput.schema.json";
import UpdatePlaceInputSchema from "~/forms/UpdatePlaceInput.schema.json";
import UpdateTransitionInputSchema from "~/forms/UpdateTransitionInput.schema.json";
import type { FormInput } from "~/lib/components/formInput";

export enum ActionKind {
  CREATE = "add",
  UPDATE = "update",
  DELETE = "delete",
}

export enum InputKind {
  PLACE = "place",
  TRANSITION = "transition",
  ARC = "arc",
}

type ActionParams = {
  kind: ActionKind,
  inputKind: InputKind,
  action: Function,
  requiredFields: string[],
  input?: FormInput,
  schema: JsonSchema,
  ui?: UISchemaElement,
  buttons: {
    label: string,
    route?: string,
    color?: string,
  }[]
}

export const defaultActions: ActionParams[] = [
  {
    kind: ActionKind.CREATE,
    inputKind: InputKind.PLACE,
    action: addPlace,
    requiredFields: ["name", "description", "bound"],
    schema: PlaceInputSchema,
    buttons: [{
      label: "Add Place"
    }],
    input: {
      name: "",
      description: "",
      bound: 1
    }
  },
  {
    kind: ActionKind.UPDATE,
    inputKind: InputKind.PLACE,
    action: updatePlace,
    requiredFields: ["id", "name", "description", "bound", "inputs", "outputs"],
    schema: UpdatePlaceInputSchema,
    buttons: [{
      label: "Update Place"
    },
      {
        label: "Delete Place",
        route: "delete-place",
        color: "bg-red-500"
      }]
  },
  {
    kind: ActionKind.DELETE,
    inputKind: InputKind.PLACE,
    action: deletePlace,
    requiredFields: ["id"],
    schema: UpdatePlaceInputSchema,
    buttons: [{
      label: "Delete Place",
      color: "bg-red-500"
    }]
  },
  {
    kind: ActionKind.CREATE,
    inputKind: InputKind.TRANSITION,
    action: addTransition,
    requiredFields: ["name", "description"],
    schema: TransitionInputSchema,
    buttons: [{
      label: "Add transition"
    }],
    input: {
      name: "",
      description: ""
    }
  },
  {
    kind: ActionKind.UPDATE,
    inputKind: InputKind.TRANSITION,
    action: updateTransition,
    schema: UpdateTransitionInputSchema,
    requiredFields: ["id", "name", "description", "inputs", "outputs"],
    buttons: [
      {
        label: "Update transition",
        route: "update-transition"
      },
      {
        label: "Delete transition",
        route: "delete-transition"
      }]
  },
  {
    kind: ActionKind.DELETE,
    inputKind: InputKind.TRANSITION,
    action: deleteTransition,
    schema: UpdateTransitionInputSchema,
    requiredFields: ["id"],
    buttons: [{
      label: "Delete transition",
      route: "delete-transition"
    }]
  },
  {
    kind: ActionKind.CREATE,
    inputKind: InputKind.ARC,
    action: addArc,
    schema: ArcInputSchema,
    requiredFields: ["placeID", "transitionID", "fromPlace"],
    buttons: [{
      label: "Add arc"

    }],
    input: {
      placeID: "",
      transitionID: "",
      fromPlace: false
    }
  },
  {
    kind: ActionKind.DELETE,
    inputKind: InputKind.ARC,
    action: deleteArc,
    schema: ArcInputSchema,
    requiredFields: ["id"],
    buttons: [{
      label: "Delete arc"
    }]
  }
];
