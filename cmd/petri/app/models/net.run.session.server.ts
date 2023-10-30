import { z } from "zod";
import { prisma } from "~/db.server";
import type {
  Instance,
  SessionState,
  Constant,
  Device,
  Field,
  Event,
  Session,
  User,
  Datum,
  Run,
  Action
} from "@prisma/client";
import { RunDetails } from "./net.run.server";

export const RunSessionCreateInputSchema = z.object({
  runID: z.string().cuid(),
  userID: z.string().cuid()
});

export type RunSessionCreateInput = z.infer<typeof RunSessionCreateInputSchema>;

export type RunSessionCreateResponse = Pick<Session, "id">

export async function createSession(input: RunSessionCreateInput): Promise<RunSessionCreateResponse> {
  const { runID, userID } = RunSessionCreateInputSchema.parse(input);
  return await prisma.session.create({
    data: {
      run: {
        connect: {
          id: runID
        }
      },
      user: {
        connect: {
          id: userID
        }
      }
    }
  });
}

export const GetRunSessionInputSchema = z.object({
  id: z.string().cuid()
});

export type GetRunSessionInput = z.infer<typeof GetRunSessionInputSchema>;

export type DataListItem = Pick<Datum, "instanceID" | "value"> & {
  createdAt: Date | string
}


export type RunSessionDetails = Pick<Session, "id"> & {
  createdAt: Date | string | null,
  updatedAt: Date | string | null,
  state: SessionState,
  startedAt: Date | string | null,
  stoppedAt: Date | string | null,
  pausedAt: (Date | string)[],
  resumedAt: (Date | string)[],
  user: Pick<User, "email">,
  run: RunDetails,
  instances: (Pick<Instance, "name" | "id"> & { device: Pick<Device, "id" | "name"> })[]
  data: DataListItem[];
}

export async function getRunSession(input: GetRunSessionInput): Promise<RunSessionDetails> {
  const { id } = GetRunSessionInputSchema.parse(input);
  return prisma.session.findUniqueOrThrow({
    where: {
      id,
      deleted: false
    },
    include: {
      user: {
        select: {
          email: true
        }
      },
      instances: {
        select: {
          id: true,
          name: true,
          device: {
            select: {
              id: true,
              name: true
            }
          }
        }
      },
      run: {
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
                  deviceId: true,
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
                  device: {
                    select: {
                      id: true,
                      name: true
                    }
                  },
                  constants: {
                    select: {
                      id: true,
                      field: {
                        select: {
                          id: true,
                          name: true,
                          type: true
                        }
                      },
                      value: true
                    }
                  }
                }
              }
            }
          }
        }
      },
      data: {
        select: {
          instanceID: true,
          value: true,
          createdAt: true
        }
      }
    }
  });
}

export const ListRunSessionsInputSchema = z.object({
  runID: z.string().cuid(),
  take: z.number().int().positive().default(10),
  offset: z.number().int().default(0)
});

export type ListRunSessionsInput = z.infer<typeof ListRunSessionsInputSchema>;

export type RunSessionListItem = Pick<Session, "id" | "createdAt" | "updatedAt">

export type ListRunSessionsResponse = {
  sessions: RunSessionListItem[],
  count: number
}

export async function listRunSessions(input: ListRunSessionsInput): Promise<ListRunSessionsResponse> {
  const { runID, take, offset } = ListRunSessionsInputSchema.parse(input);
  const sessions = await prisma.session.findMany({
    where: {
      runID,
      deleted: false
    },
    take,
    skip: offset
  });
  const countResult = await prisma.session.count({
    where: {
      runID
    }
  });
  return {
    sessions,
    count: countResult
  };
}

export const DeleteRunSessionInputSchema = z.object({
  id: z.string().cuid()
});

export type DeleteRunSessionInput = z.infer<typeof DeleteRunSessionInputSchema>;

export type DeleteRunSessionResponse = Pick<Session, "id">

export async function deleteRunSession(input: DeleteRunSessionInput): Promise<DeleteRunSessionResponse> {
  const { id } = DeleteRunSessionInputSchema.parse(input);
  return await prisma.session.update({
    where: {
      id
    },
    data: {
      deleted: true
    }
  });
}