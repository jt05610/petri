import type { LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import { sayHello } from "~/models/control.server";
import { requireUserId } from "~/session.server";
import { getUserById } from "~/models/user.server";
import { badRequest } from "~/util/request.server";

export const action = async ({ request }: LoaderArgs) => {
  const userID = await requireUserId(request);
  const req = await request.json();
  const user = await getUserById(userID);
  if (!user) {
    return badRequest("User not found");
  }
  const ok = await sayHello({ deviceID: req.deviceID });
  return json({ success: ok }, { status: 200 });
};