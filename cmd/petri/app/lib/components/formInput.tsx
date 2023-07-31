import type {PlaceInput} from "~/models/place.server";
import type {TransitionInput} from "~/models/transition.server";
import type {ArcInput} from "~/models/arc.server";
import type { NodeUpdate } from "~/lib/components/node";

export type FormInput = PlaceInput | TransitionInput | ArcInput | NodeUpdate | {
  id: string,

};