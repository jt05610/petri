import type { ActionArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { addArc } from "~/models/net.server";
import invariant from "tiny-invariant";

export const action = async ({ params, request }: ActionArgs) => {
  const formData = await request.formData();
  invariant(params.netID, "netID not found");
  const netID = params.netID;
  const placeID = formData.get("placeID");
  const transitionID = formData.get("transitionID");
  const fromPlace = formData.get("fromPlace");

  if (typeof placeID !== "string" || placeID.length === 0) {
    return json(
      { errors: { body: null, title: "PlaceID is required" } },
      { status: 400 }
    );
  }
  if (typeof transitionID !== "string" || transitionID.length === 0) {
    return json(
      { errors: { body: null, title: "TransitionID is required" } },
      { status: 400 }
    );
  }
  if (typeof fromPlace !== "string" || fromPlace.length === 0) {
    return json(
      { errors: { body: null, title: "FromPlace is required" } },
      { status: 400 }
    );
  }
  const res = await addArc({
    placeID: placeID,
    transitionID: transitionID,
    fromPlace: Boolean(fromPlace),
    netID: netID
  });
  return json(res);
};