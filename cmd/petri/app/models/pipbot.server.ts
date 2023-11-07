import invariant from "tiny-invariant";
import type { PipBotDesignInput } from "~/models/pipbot";
import { PipBotDesignInputSchema } from "~/models/pipbot";

invariant(process.env.PIPBOT_URL, "Experiment URL must be set");

const PIPBOT_URL = process.env.PIPBOT_URL;

export async function makeTransfers(input: PipBotDesignInput) {
  const design = PipBotDesignInputSchema.safeParse(input);
  invariant(design.success, "Invalid experiment design");
  console.log(design);
  return await fetch(PIPBOT_URL, {
    method: "POST",
    body: JSON.stringify(design.data),
    headers: {
      "Content-Type": "application/json"
    }
  }).then(res => res.json());
}
