import type { Prisma } from "@prisma/client";
import { prisma } from "~/db.server";
import type {
  AddActionToRunInput,
  AddConstantToAction, AddStepsToRunInput, ConstantDetails,
  GetRunInput,
  RemoveConstantInput,
  RunDetails,
  RunInput,
  RunListInput, RunListItem,
  UpdateConstantInput,
  UpdateRunInput
} from "./net.run";
import {
  AddActionToRunSchema,
  AddConstantToActionSchema,
  getRunInputSchema, RemoveConstantSchema,
  RunInputSchema, RunListInputSchema,
  UpdateConstantSchema,
  updateRunInputSchema
} from "./net.run";

type InputJsonValue = Prisma.InputJsonValue;

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

export async function addRun(netID: string, input: RunInput) {
  const { name, description, parameters } = RunInputSchema.parse(input);
  return prisma.run.create({
    data: {
      name,
      description,
      parameters: {
        create: parameters
      },
      net: {
        connect: {
          id: netID
        }
      }

    }
  });
}

export async function addRunSteps(runID: string, data: AddStepsToRunInput) {
  return prisma.run.update({
    where: { id: runID },
    data: {
      steps: {
        create: data.actions.map((action, i) => ({
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

export async function updateRun(input: UpdateRunInput) {
  const { runID, name, description, parameters } = updateRunInputSchema.parse(input);
  return prisma.run.update({
    where: { id: runID },
    data: {
      name,
      description,
      parameters: parameters && {
        upsert: parameters.map((parameter) => ({
            where: { id: parameter.id },
            update: parameter,
            create: parameter
          })
        )
      }
    }
  });
}

export async function getRunDetails(input: GetRunInput): Promise<RunDetails> {
  const { runID } = getRunInputSchema.parse(input);
  return prisma.run.findUniqueOrThrow({
    where: { id: runID },
    select: {
      id: true,
      name: true,
      description: true,
      parameters: {
        select: {
          name: true,
          expression: true
        }
      },
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

