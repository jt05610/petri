import type { User, Net, Place, Transition, Arc } from "@prisma/client";
import { prisma } from "~/db.server";

export function getNet({
                         id,
                         authorID
                       }: Pick<Net, "id"> & {
  authorID: User["id"];
}) {
  return prisma.net.findFirst({
    select: {
      id: true,
      authorID: true,
      places: {
        select: {
          place: true
        }
      },
      transitions: {
        select: {
          transition: true
        }
      },
      arcs: true,
      name: true,
      createdAt: true,
      updatedAt: true,
      description: true
    },
    where: { id, authorID }
  }).then((net) => {
    if (!net) {
      throw new Error("net not found");
    }
    return {
      ...net,
      places: net.places.map((place) => place.place),
      transitions: net.transitions.map((transition) => transition.transition)
    };
  });
}

export function getNetListItems({ authorID }: { authorID: User["id"] }) {
  return prisma.net.findMany({
    where: { authorID },
    select: { id: true, authorID: true, name: true, createdAt: true, updatedAt: true },
    orderBy: { updatedAt: "desc" }
  });
}

type CreateNetInput = Pick<Net, "name" | "description"> & {
  authorID: User["id"];
}

export function createNet({ name, description, authorID }: CreateNetInput): Promise<Net> {
  return prisma.net.create({
    data: {
      name: name,
      description: description,
      author: {
        connect: {
          id: authorID
        }
      }
    }
  });
}

export type PlaceInput = Pick<Place, "name" | "bound" | "description">


export function addPlace({
                           name,
                           bound,
                           description,
                           netID
                         }: PlaceInput & { netID: Net["id"] }) {
  return prisma.place.create({
    data: {
      name: name,
      bound: bound,
      description: description,
      nets: {
        create: {
          netID: netID
        }
      }
    }
  });
}

export type TransitionInput = Pick<Transition, "name" | "description" | "condition">

export function addTransition({ name, description, condition, netID }: TransitionInput & { netID: Net["id"] }) {
  return prisma.transition.create({
    data: {
      name: name,
      description: description,
      condition: condition,
      nets: {
        create: {
          netID: netID
        }
      }
    }
  });
}

export type ArcInput = Pick<Arc, "placeID" | "transitionID" | "fromPlace">

export function addArc({ placeID, transitionID, netID, fromPlace }: ArcInput & { netID: Net["id"] }) {
  return prisma.arc.create({
    data: {
      place: {
        connect: {
          id: placeID
        }
      },
      transition: {
        connect: {
          id: transitionID
        }
      },
      net: {
        connect: {
          id: netID
        }
      },
      fromPlace: fromPlace
    }
  });
}


export function deletePlace({ id }: Pick<Place, "id">) {
  return prisma.place.deleteMany({
    where: { id }
  });
}

export function deleteTransition({ id }: Pick<Transition, "id">) {
  return prisma.transition.deleteMany({
    where: { id }
  });
}

export function deleteArc({
                            id,
                            netID
                          }: Pick<Arc, "id"> & { netID: Net["id"] }) {
  return prisma.arc.deleteMany({
    where: { id, netID }
  });
}

export function updateNet({
                            id,
                            name,
                            authorID
                          }: Net & { authorID: User["id"]; }) {
  return prisma.net.updateMany({
    where: { id, authorID },
    data: { name }
  });
}

export function updatePlace({ id, name, bound }: Place) {
  return prisma.place.updateMany({
    where: { id },
    data: { name, bound }
  });
}

export function updateTransition({ id, name }: Transition) {
  return prisma.transition.updateMany({
    where: { id },
    data: { name }
  });
}

export function deleteNet({ id, authorID }: Pick<Net, "id"> & { authorID: User["id"] }) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}
