import type { User, Net, Arc, Transition, Place, Event } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";

export const NetInputSchema = z.object({
  name: z.string(),
  authorID: z.string().cuid(),
  description: z.string()
});

export type NetInput = z.infer<typeof NetInputSchema>;

export async function createNet(input: NetInput) {
  const { name, authorID, description } = NetInputSchema.parse(input);
  return prisma.net.create({
    data: {
      authorID,
      name,
      description
    }
  });
}

export const NetUpdateSchema = z.object({
  id: z.string().cuid().optional(),
  name: z.string().optional(),
  description: z.string().optional()
});

export type NetUpdate = z.infer<typeof NetUpdateSchema>;

export async function updateNet(input: NetUpdate) {
  const { id, name, description } = NetUpdateSchema.parse(input);
  return prisma.net.update({
    where: { id },
    data: {
      name,
      description
    }
  });
}

export type TransitionWithEvents = Pick<Transition, "id" | "name"> & {
  events?: (Pick<Event, "id" | "name"> & {
    fields: {
      name: string
      type: "string" | "number" | "boolean" | string
    }[]
  })[]
};

export type NetDetails = Pick<Net, "id" | "name" | "initialMarking"> & {
  places: Pick<Place, "id" | "name" | "bound">[]
  transitions: TransitionWithEvents[]
  device: {
    instances: {
      id: string
      name: string
      addr: string
    }[] | null
  } | null
  arcs: Pick<Arc, "placeID" | "fromPlace" | "transitionID">[]
}

export async function getNet({
                               id,
                               authorID
                             }: Pick<Net, "id"> & {
  authorID: User["id"];
}) {
  return prisma.net.findFirst({
    select: {
      id: true,
      authorID: true,
      parentID: true,
      places: {
        select: {
          id: true,
          name: true,
          bound: true
        }
      },
      transitions: {
        select: {
          id: true,
          name: true
        }
      },
      arcs: true,
      name: true,
      createdAt: true,
      updatedAt: true,
      description: true,
      placeInterfaces: {
        select: {
          id: true,
          name: true,
          places: {
            select: {
              id: true,
              name: true
            }
          }
        }
      },
      transitionInterfaces: {
        select: {
          id: true,
          name: true,
          transitions: {
            select: {
              id: true,
              name: true
            }
          }
        }
      },
      initialMarking: true,
      children: {
        select: {
          id: true,
          name: true,
          description: true,
          places: {
            select: {
              id: true,
              name: true
            }
          },
          transitions: {
            select: {
              id: true,
              name: true
            }
          },
          arcs: true
        }
      }
    },
    where: { id, authorID }
  }).then((net) => {
    if (!net) {
      throw new Error("net not found");
    }
    return net;
  });
}

export type NetListItem = {
  id: Net["id"];
  authorID: Net["authorID"];
  name: Net["name"];
  createdAt: string;
  updatedAt: string;
}

export async function getNetListItems({ authorID }: { authorID: User["id"] }) {
  return prisma.net.findMany({
    where: { authorID },
    select: { id: true, authorID: true, name: true, createdAt: true, updatedAt: true },
    orderBy: { updatedAt: "desc" }
  });
}


export async function getNetWithDeviceInstances({ id, authorID }: Pick<Net, "id"> & {
  authorID: User["id"]
}) {
  return prisma.net.findFirst({
    where: { id, authorID },
    select: {
      id: true,
      name: true,
      initialMarking: true,
      places: {
        select: {
          id: true,
          name: true,
          bound: true
        }
      },
      transitions: {
        select: {
          id: true,
          name: true,
          events: {
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
          }
        }
      },
      arcs: {
        select: {
          id: true,
          fromPlace: true,
          placeID: true,
          transitionID: true
        }
      },
      device: {
        select: {
          instances: {
            select: {
              id: true,
              addr: true,
              name: true
            }
          }
        }
      }
    }

  }).then((net) => {
    if (!net) {
      throw new Error("net not found");
    }
    return net;
  });
}

export function deleteNet({ id, authorID }: Pick<Net, "id"> & { authorID: User["id"] }) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}
