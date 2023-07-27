import type { User, Net, Place, Transition, Arc } from "@prisma/client";
import { prisma } from "~/db.server";

export function getNet({
                          id,
                          authorID,
                        }: Pick<Net, "id"> & {
  authorID: User["id"];
}) {
  return prisma.net.findFirst({
    select: {
      id: true,
      authorID: true,
      places: true,
      transitions: true,
      arcs: true,
      name: true,
      createdAt: true,
      updatedAt: true,
      description: true
    },
    where: { id, authorID }
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

export function addPlace({
                           name,
                           bound,
                           netID
                         }: Pick<Place, "name" | "bound"> & {
  netID: Net["id"];
}) {
  return prisma.place.create({
    data: {
      name: name,
      bound: bound,
      net: {
        connect: {
          id: netID
        }
      }
    }
  });
}

export function addTransition(
  {
    name,
    netID
  }: Transition & {
    netID: Net["id"];
  }) {
  return prisma.transition.create({
    data: {
      name: name,
      net: {
        connect: {
          id: netID
        }
      }
    }
  });
}

export function addArc({
                         placeID,
                         transitionID,
                         netID,
                         fromPlace
                       }: Arc & {
  netID: Net["id"];
}) {
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


export function deletePlace({
                              id,
                              netID
                            }: Pick<Place, "id"> & { netID: Net["id"] }) {
  return prisma.place.deleteMany({
    where: { id, netID }
  });
}

export function deleteTransition({
                                   id,
                                   netID
                                 }: Pick<Transition, "id"> & { netID: Net["id"] }) {
  return prisma.transition.deleteMany({
    where: { id, netID }
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

export function updatePlace({
                              id,
                              name,
                              bound,
                              netID
                            }: Place & { netID: User["id"]; }) {
  return prisma.place.updateMany({
    where: { id, netID },
    data: { name, bound }
  });
}

export function updateTransition({
                                   id,
                                   name,
                                   netID
                                 }: Transition & { netID: User["id"]; }) {
  return prisma.transition.updateMany({
    where: { id, netID },
    data: { name }
  });
}

export function deleteNet({
                            id,
                            authorID
                          }: Pick<Net, "id"> & { authorID: User["id"] }) {
  return prisma.net.deleteMany({
    where: { id, authorID }
  });
}
