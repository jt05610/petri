import { z } from "zod";
import type { Parameter } from "~/models/parameter";
import { RunParameterSchema, UpdateParameterSchema } from "~/models/parameter";
import { prisma } from "~/db.server";
import type { Step, Action, Constant, Field, Run } from "@prisma/client";


// Device event constant CRUD operations
export const ConstantInputSchema = z.object({
  fieldID: z.string().cuid(),
  value: z.string()
});

// Device event constant CRUD operations
export const ConstantInputDisplaySchema = z.object({
  fieldID: z.string().cuid(),
  constant: z.boolean(),
  value: z.string()
});

export type ConstantInputDisplay = z.infer<typeof ConstantInputDisplaySchema>;

export type ConstantInput = z.infer<typeof ConstantInputSchema>

export const AddConstantToActionSchema = ConstantInputSchema.extend({
  actionId: z.string().cuid()
});

export type AddConstantToAction = z.infer<typeof AddConstantToActionSchema>;


export const UpdateConstantSchema = z.object({
  actionId: z.string().cuid(),
  constantId: z.string().cuid(),
  value: z.string()
});

export type UpdateConstantInput = z.infer<typeof UpdateConstantSchema>;


export const RemoveConstantSchema = z.object({
  actionId: z.string().cuid(),
  constantId: z.string().cuid()
});

export type RemoveConstantInput = z.infer<typeof RemoveConstantSchema>;


const literalSchema = z.union([z.string(), z.number(), z.boolean(), z.null()]);
type Literal = z.infer<typeof literalSchema>;
type Json = Literal | {
  [key: string]: Json
} | Json[];
export const JSONSchema: z.ZodType<Json> = z.lazy(() =>
  z.union([literalSchema, z.array(JSONSchema), z.record(JSONSchema)])
);


// Device event CRUD operations

export const ActionInputSchema = z.object({
  deviceId: z.string().cuid(),
  input: z.any(),
  output: z.any(),
  eventID: z.string().cuid(),
  eventName: z.string(),
  constants: z.array(ConstantInputSchema),
  instruction: z.string().optional(),
  documentation: z.string().optional()
});

export const ActionInputDisplaySchema = z.object({
    deviceId: z.string().cuid(),
    eventID: z.string().cuid(),
    eventFields: z.array(z.object({
      id: z.string().cuid(),
      name: z.string(),
      type: z.string()
    })),
    input: z.any(),
    output: z.any(),
    eventName: z.string(),
    constants: z.array(ConstantInputDisplaySchema),
    instruction: z.string().optional(),
    documentation: z.string().optional()
  }
);

export type ActionInputDisplay = z.infer<typeof ActionInputDisplaySchema>;

export type ActionInput = z.infer<typeof ActionInputSchema>;

export const AddActionToRunSchema = ActionInputSchema.extend({
  runId: z.string().cuid()
});

export type AddActionToRunInput = z.infer<typeof AddActionToRunSchema>;

export const RemoveActionFromRunInputSchema = z.object({
  runID: z.string().cuid(),
  actionId: z.string().cuid()
});

export type RemoveActionFromRun = z.infer<typeof RemoveActionFromRunInputSchema>;

// Run CRUD operations

export const RunInputSchema = z.object({
  name: z.string(),
  description: z.string(),
  parameters: z.array(RunParameterSchema)
});

export const AddStepsToRunSchema = z.object({
  actions: z.array(ActionInputSchema)
});

export type AddStepsToRunInput = z.infer<typeof AddStepsToRunSchema>;

export const RunInputDisplaySchema = z.object({
  name: z.string(),
  description: z.string(),
  netID: z.string().cuid(),
  deviceNames: z.array(z.string()),
  parameters: z.array(RunParameterSchema),
  actions: z.array(ActionInputDisplaySchema)
});

export type RunInputDisplay = z.infer<typeof RunInputDisplaySchema>;

export type RunInput = z.infer<typeof RunInputSchema>;


export const updateRunInputSchema = z.object({
  runID: z.string().cuid(),
  name: z.string().optional(),
  description: z.string().optional(),
  parameters: z.array(UpdateParameterSchema).optional()
});

export type UpdateRunInput = z.infer<typeof updateRunInputSchema>;


export const removeRunInputSchema = z.object({
  runID: z.string().cuid()
});

export type RemoveRunInput = z.infer<typeof removeRunInputSchema>;

export async function removeRun(input: RemoveRunInput) {
  const { runID } = removeRunInputSchema.parse(input);
  return prisma.run.delete({
    where: { id: runID }
  });
}

export const getRunInputSchema = z.object({
  runID: z.string().cuid()
});

export type GetRunInput = z.infer<typeof getRunInputSchema>;

export type ConstantDetails = Pick<Constant, "id" | "fieldID" | "value"> & {
  field: Pick<Field, "id" | "name" | "type">
}

export type ActionDetails = Pick<Action, "id"> & {
  event: {
    id: string
    name: string
    description: string | null
    fields: Pick<Field, "id" | "name" | "type">[]
  }
  device: {
    id: string,
    name: string
    instances: {
      name: string
      addr: string
    }[]
  }
  input?: any
  output?: any
  constants: ConstantDetails[]
}

export type RunDetails = Pick<Run, "id" | "name" | "description"> & {
  steps: (Pick<Step, "id" | "order"> & {
    action: ActionDetails
  })[]
  parameters: Pick<Parameter, "name" | "expression">[]
}


export const RunListInputSchema = z.object({
  netID: z.string().cuid()
});

export type RunListInput = z.infer<typeof RunListInputSchema>;

export type RunListItem = Pick<Run, "id" | "name" | "description">;




