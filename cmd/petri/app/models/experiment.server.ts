import invariant from "tiny-invariant";
import type { ExperimentDesignType} from "~/models/experiment";
import { ExperimentSchema } from "~/models/experiment";

invariant(process.env.EXPERIMENT_URL, "Experiment URL must be set");

const EXPERIMENT_URL = process.env.EXPERIMENT_URL;

export type ExperimentSample = {
  id: string;
  name: string;
  group: string;
} & Record<string, number | string>

export type ExperimentDesignResult = ExperimentSample[]

export async function makeExperiment(input: ExperimentDesignType): Promise<ExperimentDesignResult> {
  const design = ExperimentSchema.safeParse(input);
  invariant(design.success, "Invalid experiment design");
  return await fetch(EXPERIMENT_URL, {
    method: "POST",
    body: JSON.stringify(design),
    headers: {
      "Content-Type": "application/json"
    }
  }).then(res => res.json() as Promise<ExperimentDesignResult>);
}
