import invariant from "tiny-invariant";
import { z } from "zod";

export enum ExperimentType {
  lhs = "lhs",
  fullFactorial = "full",
  boxBehnken = "box",
  plackettBurman = "plackett",
}

invariant(process.env.EXPERIMENT_URL, "Experiment URL must be set");

const EXPERIMENT_URL = process.env.EXPERIMENT_URL;

interface ExperimentDimension {
  name: string;
}

export interface NumberDimension extends ExperimentDimension {
  axis: {
    min: number;
    max: number;
    step: number;
  };
}

export interface CategoricalDimension extends ExperimentDimension {
  values: (string | number)[];
}

interface ExperimentDesign {
  kind: ExperimentType;
  axes: ExperimentDimension[];
}

interface LHSDesign extends ExperimentDesign {
  kind: ExperimentType.lhs;
  n_samples: number;
  id_stem: string;
  strength: number;
}

export function makeLHSDesign(n_samples: number, id_stem: string, strength: number, axes: ExperimentDimension[]): LHSDesign {
  return {
    kind: ExperimentType.lhs,
    n_samples,
    id_stem,
    strength,
    axes
  };
}

const numberAxisSchema = z.object({
  min: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number()),
  max: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number()),
  step: z.preprocess((x) => {
    if (typeof x === "string") {
      return parseFloat(x);
    }
    return x;
  }, z.number())
});

const categoricalAxisSchema = z.array(z.string());

const axisSchema = z.object({
  name: z.string(),
  values: categoricalAxisSchema.optional(),
  axis: numberAxisSchema.optional()
});

const schema = z.object({
  kind: z.nativeEnum(ExperimentType),
  axes: z.array(axisSchema),
  n_samples: z.preprocess((x) => {
      if (typeof x === "string") {
        return parseInt(x);
      }
      return x;
    },
    z.number().int().positive()),
  strength: z.preprocess((x) => {
      if (typeof x === "string") {
        return parseInt(x);
      }
      return x;
    }
    , z.number().int().positive()),
  id_stem: z.string().optional()
});

export type ExperimentSample = {
  id: string;
  name: string;
  group: string;
} & Record<string, number | string>

export type ExperimentDesignResult = ExperimentSample[]

export async function makeExperiment(design: z.infer<typeof schema>): Promise<ExperimentDesignResult> {
  return await fetch(EXPERIMENT_URL, {
    method: "POST",
    body: JSON.stringify(schema.parse(design)),
    headers: {
      "Content-Type": "application/json"
    }
  }).then(res => res.json() as Promise<ExperimentDesignResult>);
}
