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

export type EventDetails = Pick<Event, "id" | "name"> & {
  fields: {
    id: string
    name: string
    type: "string" | "number" | "boolean" | string
  }[]
}

export type EventDetailsWithEnabled = EventDetails & {
  enabled: boolean
}

export type TransitionWithEvents = Pick<Transition, "id" | "name"> & {
  events?: EventDetails[]
};

export type NetDetails = Pick<Net, "id" | "name" | "description" | "initialMarking"> & {
  places: Pick<Place, "id" | "name" | "bound">[]
  placeInterfaces: {
    id: string
    name: string
    bound: number
    places: {
      id: string
    }[]
  }[]
  transitionInterfaces: (TransitionWithEvents & {
    transitions: {
      id: string
    }[]
  })[]
  transitions: TransitionWithEvents[]
  device: {
    id: string
    name: string
    instances: {
      id: string
      name: string
      addr: string
    }[] | null
  } | null
  arcs: Pick<Arc, "placeID" | "fromPlace" | "transitionID">[]
}

export type NetDetailsWithChildren = NetDetails & {
  children: NetDetails[]
}

export async function getNet({
                               id,
                               authorID
                             }: Pick<Net, "id"> & {
  authorID: User["id"];
}): Promise<NetDetailsWithChildren> {

  const select = {
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
      arcs: true,
      name: true,
      createdAt: true,
      updatedAt: true,
      description: true,
      device: {
        select: {
          id: true,
          name: true,
          instances: {
            select: {
              id: true,
              name: true,
              addr: true
            }
          }
        }
      },
      placeInterfaces: {
        select: {
          id: true,
          name: true,
          bound: true,
          places: {
            select: {
              id: true
            }
          }
        }
      },
      transitionInterfaces: {
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
          },
          transitions: {
            select: {
              id: true
            }
          }
        }
      },
      initialMarking: true
    }
  };

  const netSelectWithChildren = {
    ...select,
    select: {
      ...select.select,
      children: {
        select: select.select
      }
    }
  };

  return prisma.net.findFirstOrThrow({
    where: { id, authorID },
    ...netSelectWithChildren
  });
}

export type NetListItem = {
  id: Net["id"];
  authorID: Net["authorID"];
  name: Net["name"];
  createdAt: Date;
  updatedAt: Date;
}

export async function getNetListItems({ authorID }: {
  authorID: User["id"]
}): Promise<NetListItem[]> {
  return prisma.net.findMany({
    where: { authorID },
    select: { id: true, authorID: true, name: true, createdAt: true, updatedAt: true },
    orderBy: { updatedAt: "desc" }
  });
}

export function getNetsWithEvents({ authorID }: {
  authorID: User["id"]
}) {
  return prisma.net.findMany({
    where: {
      authorID,
      children: {
        some: {
          transitions: {
            some: {
              events: {
                some: {}
              }
            }
          }
        }
      }
    },
    select: {
      id: true,
      name: true,
      description: true
    }
  });
}

export function getNetsWithDevice({ authorID }: {
  authorID: User["id"]
}) {
  return prisma.net.findMany({
    where: {
      authorID,
      device: {
        isNot: null
      }
    },
    select: {
      id: true,
      name: true,
      description: true
    }
  });
}

export function deleteNet({ id, authorID }: Pick<Net, "id"> & {
  authorID: User["id"]
}) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}
