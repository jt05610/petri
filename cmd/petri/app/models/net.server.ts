import type {
  Arc,
  Transition,
  Place,
  Event,
  Net,
  User,
  SharedNet
} from "@prisma/client";
import {
  Visibility
} from "@prisma/client";
import { prisma } from "~/db.server";
import type { NetInput, UpdateNet } from "~/models/net";
import { NetInputSchema, UpdateNetSchema } from "~/models/net";

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

export async function updateNet(id: string, input: UpdateNet) {
  const { name, description, visibility, sharedWith } = UpdateNetSchema.parse(input);
  return prisma.net.update({
    where: { id },
    data: {
      name,
      description,
      visibility,
      sharedWith: {
        deleteMany: {},
        create: sharedWith?.map(email => ({
          user: {
            connect: {
              email
            }
          }
        })) ?? []
      }
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

export type NetDetails = Pick<Net, "id" | "name" | "description" | "initialMarking" | "authorID"> & {
  places: Pick<Place, "id" | "name" | "bound">[],
  visibility: Net["visibility"],
  sharedWith: {
    user: Pick<User, "id" | "email">
  }[],
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
      initialMarking: true,
      visibility: true,
      sharedWith: {
        select: {
          user: {
            select: {
              id: true,
              email: true
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

  const net = await prisma.net.findFirstOrThrow({
    where: { id },
    ...netSelectWithChildren
  });
  if (net.visibility === Visibility.PRIVATE && net.authorID !== authorID) {
    const sharedWith = net.sharedWith.map(({ user }) => user.id);
    if (!sharedWith.includes(authorID)) {
      throw new Error("Not authorized");
    }
  }
  return net;
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

export async function getNetsWithEvents({ authorID }: {
  authorID: User["id"]
}) {
  return prisma.net.findMany({
    where: {
      OR: [
        {
          visibility: Visibility.PUBLIC
        },
        {
          sharedWith: {
            some: {
              userID: authorID
            }
          }
        },
        {
          authorID: authorID
        }
      ],
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

export async function getNetsWithDevice({ authorID }: {
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

export async function deleteNet({ id, authorID }: Pick<Net, "id"> & {
  authorID: User["id"]
}) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}

export async function shareNetWithUser({ id, authorEmail }: Pick<Net, "id"> & {
  authorEmail: User["email"]
}) {
  return prisma.net.update({
    where: { id },
    data: {
      sharedWith: {
        create: {
          user: {
            connect: {
              email: authorEmail
            }
          }
        }
      }
    }
  });
}

export async function unshareNet({ id, shareID }: Pick<Net, "id"> & {
  shareID: SharedNet["id"]
}) {
  return prisma.net.update({
    where: { id },
    data: {
      sharedWith: {
        delete: {
          id: shareID
        }
      }
    }
  });
}

export async function getSharedNets({ userID }: {
  userID: User["id"]
}) {
  return prisma.net.findMany({
    where: {
      sharedWith: {
        some: {
          userID
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

export async function makeNetPublic({ id, authorID }: Pick<Net, "id"> & {
  authorID: User["id"]
}) {
  const net = await prisma.net.findFirst({
    where: { id, authorID }
  });
  if (!net) throw new Error("Net not found");
  if (net.authorID !== authorID) throw new Error("Not authorized");
  return prisma.net.update({
    where: { id },
    data: {
      visibility: Visibility.PUBLIC
    }
  });
}

export async function getPublicNets() {
  return prisma.net.findMany({
    where: {
      visibility: Visibility.PUBLIC
    },
    select: {
      id: true,
      name: true,
      description: true
    }
  });
}

export async function getNetsVisibleToUser({ userID }: {
  userID: User["id"]
}) {
  return prisma.net.findMany({
    where: {
      OR: [
        {
          visibility: Visibility.PUBLIC
        },
        {
          sharedWith: {
            some: {
              userID
            }
          }
        },
        {
          authorID: userID
        }
      ]
    },
    select: {
      id: true,
      name: true,
      description: true,
      authorID: true,
      createdAt: true,
      updatedAt: true
    }
  });
}