import type { FieldRole, Type, User, Net, Arc, Transition, Place, Event } from "@prisma/client";
import { prisma } from "~/db.server";
import { z } from "zod";
import { createId } from "@paralleldrive/cuid2";

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
      id: createId(),
      authorID,
      name,
      description
    }
  });
}

export const NetUpdateSchema = z.object({
  id: z.string().cuid2().optional(),
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

export type FieldDetails = {
  id: string
  name: string
  type: TypeDetails
  required: boolean
  role: FieldRole
}

export type TypeDetails = Pick<Type, "id" | "name" | "description" | "deletable" | "scope"> & {
  fields: Pick<FieldDetails, "id" | "name">[]
}

export type EventDetails = Pick<Event, "id" | "name"> & {
  fields: (Pick<FieldDetails, "id" | "name" | "required" | "role"> & {
    type: TypeDetails
  })[]
}

export type EventDetailsWithEnabled = EventDetails & {
  enabled: boolean
}

export type TransitionWithEvents = Pick<Transition, "id" | "name"> & {
  events?: EventDetails[]
};

export type NetDetails = Pick<Net, "id" | "name" | "description"> & {
  places: Pick<Place, "id" | "name" | "bound" | "initialMarking">[]
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
  devices: {
    device: {
      id: string
      name: string
      instances: {
        id: string
        name: string
        addr: string
      }[] | null
    }
  }[] | null
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
          bound: true,
          initialMarking: true
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
                  role: true,
                  type: {
                    select: {
                      id: true,
                      name: true,
                      description: true,
                      scope: true,
                      deletable: true,
                      fields: {
                        select: {
                          id: true,
                          name: true,
                          required: true,
                          role: true,
                          type: {
                            select: {
                              id: true,
                              name: true
                            }
                          }
                        }
                      }
                    }
                  },
                  required: true
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
      devices: {
        select: {
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
              },
              fields: {
                select: {
                  id: true,
                  name: true,
                  type: {
                    select: {
                      id: true,
                      name: true,
                      description: true,
                      scope: true,
                      deletable: true
                    }
                  },
                  required: true,
                  role: true
                }
              }
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
                  role: true,
                  type: {
                    select: {
                      id: true,
                      name: true,
                      description: true,
                      scope: true,
                      deletable: true,
                      fields: {
                        select: {
                          id: true,
                          name: true,
                          required: true,
                          role: true,
                          type: {
                            select: {
                              id: true,
                              name: true,
                              description: true,
                              scope: true
                            }
                          }
                        }
                      }
                    }
                  },
                  required: true
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
      }
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
  createdAt: Date | string;
  updatedAt: Date | string;
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
      devices: {
        some: {}
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
