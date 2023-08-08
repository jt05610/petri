import type { Prisma, Run, Constant, Action, Field } from "@prisma/client";
import { z } from "zod";
import { prisma } from "~/db.server";

type InputJsonValue = Prisma.InputJsonValue;

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

export async function addConstant(input: AddConstantToAction): Promise<ConstantDetails> {
  const { fieldID, value, actionId } = AddConstantToActionSchema.parse(input);
  return prisma.constant.create({
    data: {
      value,
      field: {
        connect: {
          id: fieldID
        }
      },
      action: {
        connect: {
          id: actionId
        }
      }
    }
  });
}

export const UpdateConstantSchema = z.object({
  actionId: z.string().cuid(),
  constantId: z.string().cuid(),
  value: z.string()
});

export type UpdateConstantInput = z.infer<typeof UpdateConstantSchema>;

export async function updateConstant(input: UpdateConstantInput) {
  const { constantId, actionId, value } = UpdateConstantSchema.parse(input);
  return prisma.action.update({
    where: { id: actionId },
    data: {
      constants: {
        update: {
          where: { id: constantId },
          data: {
            value
          }
        }
      }
    }
  });
}

export const RemoveConstantSchema = z.object({
  actionId: z.string().cuid(),
  constantId: z.string().cuid()
});

export type RemoveConstantInput = z.infer<typeof RemoveConstantSchema>;

export async function removeConstant(input: RemoveConstantInput) {
  const { constantId, actionId } = RemoveConstantSchema.parse(input);
  return prisma.action.update({
    where: { id: actionId },
    data: {
      constants: {
        delete: {
          id: constantId
        }
      }
    }
  });
}

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
  constants: z.array(ConstantInputSchema)
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
    constants: z.array(ConstantInputDisplaySchema)
  }
);

export type ActionInputDisplay = z.infer<typeof ActionInputDisplaySchema>;

export type ActionInput = z.infer<typeof ActionInputSchema>;

export const AddActionToRunSchema = ActionInputSchema.extend({
  runId: z.string().cuid()
});

export type AddActionToRunInput = z.infer<typeof AddActionToRunSchema>;


export async function addActionToRun(req: AddActionToRunInput): Promise<ActionDetails> {
  const { deviceId, input, output, runId, eventID, constants } = AddActionToRunSchema.parse(req);
  return prisma.action.create({
    data: {
      runs: {
        connect: {
          id: runId
        }
      },
      device: {
        connect: {
          id: deviceId
        }
      },
      input: input as InputJsonValue,
      output: output as InputJsonValue,
      constants: {
        create: constants
      },
      event: {
        connect: {
          id: eventID
        }
      }
    },
    include: {
      constants: true,
      event: {
        include: {
          fields: true
        }
      },
      device: {
        include: {
          instances: true
        }
      }
    }
  });
}

export const RemoveActionFromRunInputSchema = z.object({
  runID: z.string().cuid(),
  actionId: z.string().cuid()
});

export type RemoveActionFromRun = z.infer<typeof RemoveActionFromRunInputSchema>;

export async function removeActionFromRun(input: RemoveActionFromRun) {
  const { runID, actionId } = RemoveActionFromRunInputSchema.parse(input);
  return prisma.run.update({
    where: { id: runID },
    data: {
      actions: {
        delete: {
          id: actionId
        }
      }
    }
  });
}

// Run event CRUD operations


// Run CRUD operations

export const RunInputSchema = z.object({
  name: z.string(),
  description: z.string(),
  netID: z.string().cuid(),
  actions: z.array(ActionInputSchema)
});

export const RunInputDisplaySchema = z.object({
  name: z.string(),
  description: z.string(),
  netID: z.string().cuid(),
  deviceNames: z.array(z.string()),
  actions: z.array(ActionInputDisplaySchema)
});

export type RunInputDisplay = z.infer<typeof RunInputDisplaySchema>;

export type RunInput = z.infer<typeof RunInputSchema>;

export async function addRun(input: RunInput) {
  const { name, description, netID, actions } = RunInputSchema.parse(input);
  return prisma.run.create({
    data: {
      name,
      description,
      net: {
        connect: {
          id: netID
        }
      },
      actions: {
        create: actions.map((action) => ({
          device: {
            connect: {
              id: action.deviceId
            }
          },
          input: action.input as InputJsonValue,
          output: action.output as InputJsonValue,
          constants: {
            create: action.constants
          },
          event: {
            connect: {
              id: action.eventID
            }
          }
        }))
      }
    }
  });
}

export const updateRunInputSchema = z.object({
  runID: z.string().cuid(),
  name: z.string().optional(),
  description: z.string().optional()
});

export type UpdateRunInput = z.infer<typeof updateRunInputSchema>;

export async function updateRun(input: UpdateRunInput) {
  const { runID, name, description } = updateRunInputSchema.parse(input);
  return prisma.run.update({
    where: { id: runID },
    data: {
      name,
      description
    }
  });
}

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

export type ConstantDetails = Pick<Constant, "id" | "fieldID" | "value">

export type ActionDetails = Pick<Action, "id"> & {
  event: {
    id: string
    name: string
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
  constants: ConstantDetails[]
}
export type RunDetails = Pick<Run, "id" | "name" | "description"> & {
  actions: ActionDetails[]
}

export async function getRunDetails(input: GetRunInput): Promise<RunDetails> {
  const { runID } = getRunInputSchema.parse(input);
  return prisma.run.findUniqueOrThrow({
    where: { id: runID },
    select: {
      id: true,
      name: true,
      description: true,
      actions: {
        select: {
          id: true,
          device: {
            select: {
              id: true,
              name: true,
              instances: {
                select: {
                  name: true,
                  addr: true
                }
              }
            }
          },
          event: {
            select: {
              id: true,
              name: true,
              fields: {
                select: {
                  id: true,
                  name: true,
                  type: true
                }
              }
            }
          },
          constants: {
            select: {
              id: true,
              fieldID: true,
              value: true
            }
          }
        }
      }
    }
  });
}

export const RunListInputSchema = z.object({
  netID: z.string().cuid()
});

export type RunListInput = z.infer<typeof RunListInputSchema>;

export type RunListItem = Pick<Run, "id" | "name" | "description">;

export async function listRuns(input: RunListInput): Promise<RunListItem[]> {
  const { netID } = RunListInputSchema.parse(input);
  return prisma.run.findMany({
    where: { netID },
    select: {
      id: true,
      name: true,
      description: true
    }
  });
}
