import type { Step, Prisma, Run, Constant, Action, Field } from "@prisma/client";
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
    },
    select: {
      id: true,
      fieldID: true,
      value: true,
      field: {
        select: {
          id: true,
          name: true,
          type: true
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


export async function addActionToRun(req: AddActionToRunInput) {
  const { deviceId, input, output, runId, eventID, constants } = AddActionToRunSchema.parse(req);
  const actions = await prisma.step.findMany({
    where: {
      run: {
        id: runId
      }
    }
  });

  const order = actions.length;

  return prisma.run.update({
    where: { id: runId },
    data: {
      steps: {
        create: {
          order: order,
          action: {
            create: {
              input,
              output,
              device: {
                connect: {
                  id: deviceId
                }
              },
              event: {
                connect: {
                  id: eventID
                }
              },
              constants: {
                create: constants
              }
            }
          }
        }
      }
    },
    select: {
      steps: {
        where: {
          order: order
        },
        select: {
          action: {
            select: {
              id: true
            }
          }
        }
      }
    }
  }).then((run) => {
    return run.steps[0].action;
  });
}

export const RemoveActionFromRunInputSchema = z.object({
  runID: z.string().cuid(),
  actionId: z.string().cuid()
});

export type RemoveActionFromRun = z.infer<typeof RemoveActionFromRunInputSchema>;


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
      steps: {
        create: actions.map((action, i) => ({
          order: i,
          action: {
            create: {
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
  input: InputJsonValue | null
  output: InputJsonValue | null
  constants: ConstantDetails[]
}

export type RunDetails = Pick<Run, "id" | "name" | "description"> & {
  steps: (Pick<Step, "id" | "order"> & { action: ActionDetails })[]
}

export async function getRunDetails(input: GetRunInput): Promise<RunDetails> {
  const { runID } = getRunInputSchema.parse(input);
  return prisma.run.findUniqueOrThrow({
    where: { id: runID },
    select: {
      id: true,
      name: true,
      description: true,
      steps: {
        select: {
          id: true,
          order: true,
          action: {
            select: {
              id: true,
              input: true,
              output: true,
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
                  description: true,
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
                  value: true,
                  field: {
                    select: {
                      id: true,
                      name: true,
                      type: true
                    }
                  }

                }
              }
            }
          }
        },
        orderBy: {
          order: "asc"
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

